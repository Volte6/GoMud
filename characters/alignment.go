package characters

const (
	AlignmentMinimum int8 = -100
	AlignmentNeutral int8 = 0
	AlignmentMaximum int8 = 100

	AlignmentAggroThreshold int = 50 // Possible delta is 0 - 200
)

func AlignmentToString(alignment int8) string {

	if alignment < AlignmentNeutral {
		// -80 to -100
		if alignment <= AlignmentNeutral-80 {
			return `unholy`
		}
		// -60 to -79
		if alignment <= AlignmentNeutral-60 {
			return `evil`
		}
		// -40 to -59
		if alignment <= AlignmentNeutral-40 {
			return `corrupt`
		}
		// -20 to -39
		if alignment <= AlignmentNeutral-20 {
			return `misguided`
		}

	} else if alignment > AlignmentNeutral {

		// 80-100
		if alignment >= AlignmentNeutral+80 {
			return `holy`
		}
		// 60 to 79
		if alignment >= AlignmentNeutral+60 {
			return `good`
		}
		// 40 to 59
		if alignment >= AlignmentNeutral+40 {
			return `virtuous`
		}
		// 20 to 39
		if alignment >= AlignmentNeutral+20 {
			return `lawful`
		}

	}

	return `neutral`

}
