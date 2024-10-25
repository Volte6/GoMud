package gametime

var (
	monthNames = []string{
		`Arvalon`,
		`Beldris`,
		`Celmara`,
		`Durelin`,
		`Esmira`,
		`Ferulan`,
		`Glimar`,
		`Hestara`,
		`Irinel`,
		`Jorenth`,
		`Keldris`,
		`Luneth`,
	}
)

func MonthName(month int) string {
	month--
	return monthNames[month%len(monthNames)]
}
