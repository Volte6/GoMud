package characters

const (
	// MinMax alignment
	AlignmentMinimum int8 = -100
	AlignmentMaximum int8 = 100
	// Good
	AlignmentHoly     int8 = 80
	AlignmentGood     int8 = 60
	AlignmentVirtuous int8 = 40
	AlignmentLawful   int8 = 20
	// Neutral
	AlignmentNeutralHigh int8 = 19
	AlignmentNeutral     int8 = 0
	AlignmentNeutralLow  int8 = -19
	// Evil
	AlignmentMisguided int8 = -20
	AlignmentCorrupt   int8 = -40
	AlignmentEvil      int8 = -60
	AlignmentUnholy    int8 = -80
	// Threshold by which mobs will auto aggro
	AlignmentAggroThreshold int = 50 // Possible delta is 0 - 200

)

func AlignmentToString(alignment int8) string {

	if alignment < AlignmentNeutralLow {
		// -80 to -100
		if alignment <= AlignmentUnholy {
			return `unholy`
		}
		// -60 to -79
		if alignment <= AlignmentEvil {
			return `evil`
		}
		// -40 to -59
		if alignment <= AlignmentCorrupt {
			return `corrupt`
		}
		// -20 to -39
		if alignment <= AlignmentMisguided {
			return `misguided`
		}

	} else if alignment > AlignmentNeutralHigh {

		// 80-100
		if alignment >= AlignmentHoly {
			return `holy`
		}
		// 60 to 79
		if alignment >= AlignmentGood {
			return `good`
		}
		// 40 to 59
		if alignment >= AlignmentVirtuous {
			return `virtuous`
		}
		// 20 to 39
		if alignment >= AlignmentLawful {
			return `lawful`
		}

	}

	return `neutral`

}
