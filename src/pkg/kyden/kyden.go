package kyden

import "fmt"

var _ Kydener = (*Kyden)(nil)

type Kyden struct {
	name string
}

func NewKyden(name string) *Kyden {
	return &Kyden{name: name}
}

func (this *Kyden) Name() string {
	return this.name
}

func (this *Kyden) Hello() string {
	return "Hello"
}

func (this *Kyden) Say(str string) string {
	return fmt.Sprintf("%s %s", this.Hello(), str)
}
