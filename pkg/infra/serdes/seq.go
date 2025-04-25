package serdes

type Serializer interface {
	Encode(any) ([]byte, error)
	Decode([]byte, any) error
}
