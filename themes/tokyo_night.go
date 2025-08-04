package themes

var TokyoNightPalette = struct {
    // Tokyo Night specific naming
    Foreground      string
    Background      string
    BackgroundDark  string
    BackgroundFloat string
    Selection       string
    Comment         string
    Blue            string
    Blue0           string
    Blue1           string
    Blue2           string
    Blue5           string
    Blue6           string
    Blue7           string
    Cyan            string
    Green           string
    Green1          string
    Green2          string
    Magenta         string
    Magenta2        string
    Orange          string
    Purple          string
    Red             string
    Red1            string
    Teal            string
    Yellow          string
    GitAdd          string
    GitChange       string
    GitDelete       string
}{
    Foreground:      "#c0caf5",
    Background:      "#1a1b26",
    BackgroundDark:  "#16161e",
    BackgroundFloat: "#1f2335",
    Selection:       "#33467c",
    Comment:         "#565f89",
    Blue:            "#7aa2f7",
    Blue0:           "#3d59a1",
    Blue1:           "#2ac3de",
    Blue2:           "#0db9d7",
    Blue5:           "#89ddff",
    Blue6:           "#b4f9f8",
    Blue7:           "#394b70",
    Cyan:            "#7dcfff",
    Green:           "#9ece6a",
    Green1:          "#73daca",
    Green2:          "#41a6b5",
    Magenta:         "#bb9af7",
    Magenta2:        "#ff007c",
    Orange:          "#ff9e64",
    Purple:          "#9d7cd8",
    Red:             "#f7768e",
    Red1:            "#db4b4b",
    Teal:            "#1abc9c",
    Yellow:          "#e0af68",
    GitAdd:          "#449dab",
    GitChange:       "#6183bb",
    GitDelete:       "#914c54",
}

var TokyoNightScheme = ColorScheme{
    DefaultFg:   TokyoNightPalette.Foreground,
    MutedFg:     TokyoNightPalette.Comment,
    SelectedFg:  TokyoNightPalette.Cyan,
    PrimaryFg:   TokyoNightPalette.Blue,
    ErrorFg:     TokyoNightPalette.Red,
    DefaultBg:   TokyoNightPalette.Background,
    PopupBg:     TokyoNightPalette.BackgroundFloat,
    BorderColor: TokyoNightPalette.Blue7,
}