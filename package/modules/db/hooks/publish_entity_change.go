package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/common"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/db/mixin"
	"github.com/gianglt2198/federation-go/package/utils"
)

type ActionKey string

const (
	Create     ActionKey = "created"
	Update     ActionKey = "updated"
	Delete     ActionKey = "deleted"
	SoftDelete ActionKey = "soft_deleted"
)

type EntityEvent struct {
	Action  ActionKey              `json:"action"`
	Data    map[string]interface{} `json:"data"`
	OldData map[string]interface{} `json:"old_data"`
}

type DataChangeKey struct{}

func SkipTrackChange(ctx context.Context) context.Context {
	return context.WithValue(ctx, DataChangeKey{}, true)
}

func determineAction(m ent.Mutation) ActionKey {
	switch {
	case m.Op().Is(ent.OpCreate):
		return Create
	case m.Op().Is(ent.OpDelete), m.Op().Is(ent.OpDeleteOne):
		return Delete
	case m.Op().Is(ent.OpUpdate), m.Op().Is(ent.OpUpdateOne):
		for _, field := range m.Fields() {
			if field == "deleted_at" {
				return SoftDelete
			}
		}
		return Update
	default:
		return "unknown"
	}
}

func extractChangedFields(ctx context.Context, m ent.Mutation) (map[string]interface{}, map[string]interface{}) {
	data := make(map[string]interface{})
	oldData := make(map[string]interface{})
	for _, field := range m.Fields() {
		if value, ok := m.Field(field); ok {
			data[field] = value
		}
		if oldValue, err := m.OldField(ctx, field); err == nil {
			oldData[field] = oldValue
		}
	}
	return data, oldData
}

func extractEntityIDs(ctx context.Context, m ent.Mutation, entity interface{}) []string {
	val := reflect.Indirect(reflect.ValueOf(entity))
	if val.Kind() == reflect.Struct {
		idField := val.FieldByName("ID")
		if idField.IsValid() && idField.CanInterface() {
			if id, ok := idField.Interface().(string); ok {
				return []string{id}
			}
		}
	}

	if mx, ok := m.(interface{ ID() (string, bool) }); ok {
		if id, exists := mx.ID(); exists {
			return []string{id}
		}
	}

	if mx, ok := m.(interface {
		IDs(ctx context.Context) ([]string, error)
	}); ok {
		ctx = mixin.SkipSoftDelete(ctx)
		if ids, err := mx.IDs(ctx); err == nil {
			return ids
		}
	}
	return []string{}
}

func PublishEntityChangeHook(serviceName string, publisher pubsub.Publisher, log *logging.Logger) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if ctx.Value(DataChangeKey{}) == true {
				return next.Mutate(ctx, m)
			}
			data, oldData := extractChangedFields(ctx, m)
			value, err := next.Mutate(ctx, m)
			if err != nil {
				return value, err
			}

			IDs := extractEntityIDs(ctx, m, value)
			if len(IDs) > 0 {
				data["ids"] = IDs
				event := EntityEvent{
					Action:  determineAction(m),
					Data:    data,
					OldData: oldData,
				}

				publish := func(ctx context.Context) {
					eventData, err := json.Marshal(event)
					if err != nil {
						log.GetWrappedLogger(ctx).Error("Failed to marshal event data", zap.Error(err))
						return
					}

					subject := fmt.Sprintf("%s.%s.changed", strings.ToLower(serviceName), strings.ToLower(m.Type()))
					if err = publisher.Publish(ctx, subject, eventData, nil); err != nil {
						log.GetWrappedLogger(ctx).Error("Failed to send NATS message", zap.Error(err))
					} else {
						log.GetWrappedLogger(ctx).Debug("Event published to NATS subject", zap.String("subject", subject), zap.String("data", string(eventData)))
					}
				}

				requestID := utils.GetRequestIDFromCtx(ctx)
				newCtx := context.WithValue(context.Background(), common.KEY_REQUEST_ID, requestID)
				publish(newCtx)
			}
			return value, nil
		})
	}
}
