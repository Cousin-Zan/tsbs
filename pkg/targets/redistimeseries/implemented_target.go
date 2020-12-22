package redistimeseries

import (
	"github.com/blagojts/viper"
	"github.com/spf13/pflag"
	"github.com/timescale/tsbs/pkg/data/serialize"
	"github.com/timescale/tsbs/pkg/data/source"
	"github.com/timescale/tsbs/pkg/targets"
	"github.com/timescale/tsbs/pkg/targets/constants"
)

func NewTarget() targets.ImplementedTarget {
	return &redistimeseriesTarget{}
}

type redistimeseriesTarget struct {
}

func (t *redistimeseriesTarget) TargetSpecificFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	flagSet.String(flagPrefix+"host", "localhost:6379", "The host:port for Redis connection")
	pflag.Uint64(flagPrefix+"pipeline", 50, "The pipeline's size")
	pflag.Bool(flagPrefix+"compression-enabled", true, "Whether to use compressed time series")
	pflag.Bool(flagPrefix+"cluster", false, "Whether to use OSS cluster API")

}

func (t *redistimeseriesTarget) TargetName() string {
	return constants.FormatRedisTimeSeries
}

func (t *redistimeseriesTarget) Serializer() serialize.PointSerializer {
	return &Serializer{}
}

func (t *redistimeseriesTarget) Benchmark(string, *source.DataSourceConfig, *viper.Viper) (targets.Benchmark, error) {
	panic("not implemented")
}
