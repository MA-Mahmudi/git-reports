package report

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type Tday struct {
	CommitCount int
	Date        time.Time
}

type Tmonth struct {
	Month time.Month
	Tdays map[int]Tday
}

func getColor(commitCount int) string {
	if commitCount == 0 {
		return "178;215;155" // #B2D79B (Light Green)
	} else if commitCount <= 5 {
		return "139;195;74" // #8BC34A (Medium Green)
	} else if commitCount <= 10 {
		return "34;139;34" // #228B22 (Forest Green)
	} else if commitCount <= 15 {
		return "0;100;0" // #006400 (Dark Green)
	} else if commitCount <= 20 {
		return "0;128;128" // #008080 (Emerald Green)
	} else {
		return "0;64;0" // #004000 (Darker Green)
	}
}

type Tyear struct {
	Tmonths map[time.Month]Tmonth
	Year    int
}

func (y Tyear) getFirstMonth() (time.Month, error) {
	for i := 1; i < 13; i++ {
		_, ok := y.Tmonths[time.Month(i)]
		if ok {
			return time.Month(i), nil
		}
	}
	return time.Month(0), errors.New("empty name")
}

var shortDayNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var commitCountRange = []string{
	"  .0.  ",
	" *1-5* ",
	"*06-10*",
	"*11-15*",
	"*16-20*",
	" *20<* ",
}

func commitCountGuide() {
	fmt.Println()
	commitCount := 0
	pterm.DefaultBasicText.Print(pterm.Blue("commits count guide:"))
	for i := 0; i < len(commitCountRange); i++ {
		commitCount = i * 5
		color := getColor(commitCount)
		var char = commitCountRange[i]
		pterm.DefaultBasicText.Printf(" \x1b[48;2;%sm%s\x1b[0m ", color, pterm.Red(char))
	}
	fmt.Println()
}

func (y Tyear) p() {
	fmt.Println()
	newHeader := pterm.HeaderPrinter{
		TextStyle:       pterm.NewStyle(pterm.FgBlack),
		BackgroundStyle: pterm.NewStyle(pterm.BgLightGreen),
		// Margin:          20,
	}

	newHeader.WithFullWidth().Println(y.Year)
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	// border := strings.Repeat("-", width)
	// fmt.Println(border)
	monthW := 6
	offset := 1 + 5
	for {
		if (width+1-offset)%(monthW+1) == 0 {
			break
		}
		offset++
	}

	monthPerLine := (width + 1 - offset) / (monthW + 1)
	firstMonth, err := y.getFirstMonth()
	if err != nil {
		return
	}
	monthIndex := int(firstMonth)
	lineIndex := 1
	for monthIndex < 13 {
		fmt.Print("     ")
		for monthIndex-int(firstMonth)+1 <= lineIndex*monthPerLine && monthIndex < 13 {
			value, ok := y.Tmonths[time.Month(monthIndex)]
			if !ok {
				monthIndex = monthIndex + 1
				continue
			}
			fmt.Print("  ")
			pterm.DefaultBasicText.Print(pterm.Green(value.Month.String()[0:3]))
			pterm.DefaultBasicText.Print(pterm.Yellow(" |"))
			monthIndex = monthIndex + 1
		}
		fmt.Println()
		for i := 0; i < 7; i++ {
			monthIndex = monthPerLine*(lineIndex-1) + 1
			pterm.DefaultBasicText.Print(pterm.Blue(shortDayNames[i]))
			pterm.DefaultBasicText.Print(pterm.Yellow(": "))
			for monthIndex-int(firstMonth)+1 <= lineIndex*monthPerLine && monthIndex < 13 {
				value, ok := y.Tmonths[time.Month(monthIndex)]
				if !ok {
					monthIndex = monthIndex + 1
					continue
				}
				firstWeekDay := value.Tdays[1].Date.Weekday()
				for j := 1; j < 7; j++ {
					dayIndex := 7*j - int(firstWeekDay) - 6 + i
					d, ok := value.Tdays[dayIndex]
					if !ok {
						fmt.Print(" ")
						continue
					}
					color := getColor(d.CommitCount)
					var char string
					if d.CommitCount == 0 {
						char = "."
					} else {
						char = "*"
					}
					fmt.Printf("\x1b[48;2;%sm%s\x1b[0m", color, char)
				}
				pterm.DefaultBasicText.Print(pterm.Yellow("|"))
				monthIndex = monthIndex + 1
			}
			fmt.Println()
		}
		lineIndex = lineIndex + 1
	}
}

type HeatMapReport struct {
    CommitsMap map[string]int
}

func (r HeatMapReport) IterationStep(c *object.Commit)  {
    year, month, date := c.Author.When.Local().Date()
    key := fmt.Sprintf("%d-%d-%d", year, month, date)
    _, exists := r.CommitsMap[key]
    if !exists {
        r.CommitsMap[key] = 1
    } else {
        r.CommitsMap[key]++
    }
}

func (r HeatMapReport) Print(){
		if len(r.CommitsMap) == 0 {
			fmt.Println("No commits where found!")
			return
		}

		keys := make([]string, 0, len(r.CommitsMap))

		for k := range r.CommitsMap {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			timeI, _ := time.Parse("2006-1-2", keys[i])
			timeJ, _ := time.Parse("2006-1-2", keys[j])

			if timeI.Before(timeJ) {
				return true
			} else {
				return false
			}
		})

		firstDate, _ := time.Parse("2006-1-2", keys[0])
		startDate := time.Date(firstDate.Year(), firstDate.Month(), 1, 0, 0, 0, 0, firstDate.Location())

		lastDate, _ := time.Parse("2006-1-2", keys[len(keys)-1])
		endDate := time.Date(lastDate.Year(), lastDate.Month(), 1, 0, 0, 0, 0, lastDate.Location()).AddDate(0, 1, -1)

		years := make(map[int]Tyear)
		for startDate.Before(endDate) {
			_, exists := years[startDate.Year()]
			if !exists {
				years[startDate.Year()] = Tyear{Tmonths: make(map[time.Month]Tmonth), Year: startDate.Year()}
			}

			_, exists = years[startDate.Year()].Tmonths[startDate.Month()]
			if !exists {
				years[startDate.Year()].Tmonths[startDate.Month()] = Tmonth{Tdays: make(map[int]Tday), Month: startDate.Month()}
			}

			tDay, exists := years[startDate.Year()].Tmonths[startDate.Month()].Tdays[startDate.Day()]
			if !exists {
				years[startDate.Year()].Tmonths[startDate.Month()].Tdays[startDate.Day()] = Tday{}
			}
			tDay.CommitCount = r.CommitsMap[startDate.Format("2006-1-2")]
			tDay.Date = startDate

			years[startDate.Year()].Tmonths[startDate.Month()].Tdays[startDate.Day()] = tDay

			startDate = startDate.AddDate(0, 0, 1)
		}

		yearsKey := make([]int, 0, len(years))
		for k := range years {
			yearsKey = append(yearsKey, k)
		}
		sort.Ints(yearsKey)
		for _, k := range yearsKey {
			years[k].p()
		}
		commitCountGuide()
}
