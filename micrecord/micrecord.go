package micrecord

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"time"
	"strings"

	"github.com/edaniels/golog"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"
)

const defaultDuration = 10
const defaultBetween = 60

var Model = resource.ModelNamespace("beering").WithFamily("generic").WithModel("micrecord")

// Config is used for converting config attributes.
type Config struct {
	Datadir string `json:"datadir"`
	Duration int `json:"duration,omitempty"`
	Between int `json:"between,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (config *Config) Validate(path string) ([]string, error) {
	var deps []string
	if config.Datadir == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "datadir")
	}
	return deps, nil
}

func init() {
	resource.RegisterComponent(
		generic.API,
		Model,
		resource.Registration[resource.Resource, *Config]{
			Constructor: func(
				ctx context.Context,
				deps resource.Dependencies,
				conf resource.Config,
				logger golog.Logger,
			) (resource.Resource, error) {
				newConf, err := resource.NativeConfig[*Config](conf)
				if err != nil {
					return nil, err
				}
				return newMicRecord(newConf)
		},
	})
}


type micRecord struct {
	resource.Named
	resource.AlwaysRebuild
	datadir string
	duration int
	between int
	cancelFunc func()
}

func newMicRecord(attr *Config) (*micRecord, error) {
	mc := &micRecord{}

	mc.datadir = attr.Datadir
	
	duration := attr.Duration
	if duration == 0 {
		duration = defaultDuration
	}
	mc.duration = duration
	between := attr.Between
	if between == 0 {
		between = defaultBetween
	}
	mc.between = between
	
	ctx, cancelFunc := context.WithCancel(context.Background())
	mc.cancelFunc = cancelFunc
	utils.PanicCapturingGo(func() {
		for utils.SelectContextOrWait(ctx, time.Duration(mc.between)*time.Second) {
			mc.writeData()
		}
	})
	return mc, nil
}

func (mc *micRecord) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (mc *micRecord) Close(ctx context.Context) error {
	mc.cancelFunc()
	return nil
}

func (mc *micRecord) writeData() {
	recCommand := "arecord --duration=" + strconv.Itoa(mc.duration) + " -t raw -f S16_LE -r16000 | flac - -f --endian little --sign signed --channels 1 --bps 16 --sample-rate 16000 -s -c -o "
	ts := time.Now().UTC().Format(time.RFC3339)
	ts = strings.Replace(strings.Replace(strings.Replace(ts, ":", "-", -1), "T", "--", -1), "Z", "UTC", -1)
	recCommand += mc.datadir + "/" + ts + ".flac"
	fmt.Println("running", recCommand)
	err := exec.Command("bash", "-c", recCommand).Run()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
