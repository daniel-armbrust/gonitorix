//
// internal/graph/structs.go
//
package graph

type GraphTemplate struct {
	Graph         string
	Title         string
	Start         string
	VerticalLabel string
	Width         int
	Height        int
	XGrid         string
	Defs          []string
	CDefs         []string
	Draw          []string
}

type GraphPeriod struct {
    Name  string
    Start string
	XGrid string
}

var (
		Daily  = GraphPeriod{
			Name:  "daily", 
			Start: "-1day",
			XGrid: "HOUR:1:HOUR:6:HOUR:6:0:%R",
		}

		Weekly = GraphPeriod{
			Name:  "weekly", 
			Start: "-1week",
		}

		Monthly = GraphPeriod{
			Name:  "monthly", 
			Start: "-1month",
		}

		Yearly = GraphPeriod{
			Name:  "yearly", 
			Start: "-1year",
		}
)