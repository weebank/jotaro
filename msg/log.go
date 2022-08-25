package msg

import (
	"os"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

var logger = apexLogger.Logger{
	Handler: cli.New(os.Stdout),
}
