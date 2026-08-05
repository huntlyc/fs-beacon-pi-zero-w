package rgb

type RGB [3]int

type Colour int

const Red Colour = 1 << 2
const Green Colour = 1 << 1
const Blue Colour = 1 << 0
const Yellow Colour = Red | Green
const Pruple Colour = Red | Blue
const White Colour = Red | Green | Blue
const Teal Colour = Green | Blue
