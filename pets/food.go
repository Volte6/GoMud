package pets

type Food int

func (f Food) String() string {

	if f <= 0 {
		return `dying`
	}

	if f == 1 {
		return `starving`
	}

	if f == 2 {
		return `moody`
	}

	if f == 3 {
		return `hungry`
	}

	return `full`
}

func (f *Food) Add() {

	*f += 1

	if *f < 0 {
		*f = 0
	}
	if *f > 4 {
		*f = 4
	}
}

func (f *Food) Remove() {

	*f -= 1

	if *f < 0 {
		*f = 0
	}
	if *f > 4 {
		*f = 4
	}
}
