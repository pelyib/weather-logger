package business

import "fmt"

const COMMAND_FETCH_FORECASTS string = "fetch:forecasts"
const COMMAND_FETCH_HISTORICAL string = "fetch:historical"

type Command struct {
	command string
}

func (c Command) AsString() string {
	return fmt.Sprintf("command|%s", c.command)
}
