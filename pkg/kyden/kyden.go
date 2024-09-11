package kyden

import "fmt"

var _ Kydener = (*Kyden)(nil)

type Kyden struct {
	name string
}

func NewKyden(name string) *Kyden {
	return &Kyden{name: name}
}

func (k *Kyden) Name() string {
	return k.name
}

func (k *Kyden) Hello() string {
	return "Hello"
}

func (k *Kyden) Say(str string) string {
	return fmt.Sprintf("%s %s", k.Hello(), str)
}
