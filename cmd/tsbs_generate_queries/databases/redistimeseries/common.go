package redistimeseries

import (
	redistimeseries "github.com/RedisTimeSeries/redistimeseries-go"
	"github.com/timescale/tsbs/cmd/tsbs_generate_queries/uses/devops"
	"github.com/timescale/tsbs/cmd/tsbs_generate_queries/utils"
	"github.com/timescale/tsbs/query"
	"time"
)

// BaseGenerator contains settings specific for RedisTimeSeries database.
type BaseGenerator struct {
}

// GenerateEmptyQuery returns an empty query.Cassandra.
func (g *BaseGenerator) GenerateEmptyQuery() query.Query {
	return query.NewRedisTimeSeries()
}

// fill Query fills the query struct with data
func (d *BaseGenerator) fillInQueryStrings(qi query.Query, humanLabel, humanDesc string) {
	q := qi.(*query.RedisTimeSeries)
	q.HumanLabel = []byte(humanLabel)
	q.HumanDescription = []byte(humanDesc)
}

// AddQuery adds a command to be executed in the full flow of this Query
func (d *BaseGenerator) AddQuery(qi query.Query, tq [][]byte, commandname []byte) {
	q := qi.(*query.RedisTimeSeries)
	q.AddQuery(tq, commandname)
}

// SetSingleGroupByTime sets SetSingleGroupByTimestamp used for this Query
func (d *BaseGenerator) SetSingleGroupByTime(qi query.Query, value bool ) {
	q := qi.(*query.RedisTimeSeries)
	q.SetSingleGroupByTimestamp(value)
}

// SetReduceSeries sets SetReduceSeries used for this Query
func (d *BaseGenerator) SetReduceSeries(qi query.Query, value bool, reducer func(series [] redistimeseries.Range) (redistimeseries.Range, error) ) {
	q := qi.(*query.RedisTimeSeries)
	q.SetReduceSeries(value)
	q.SetReducer(reducer)
}

// NewDevops creates a new devops use case query generator.
func (g *BaseGenerator) NewDevops(start, end time.Time, scale int) (utils.QueryGenerator, error) {
	core, err := devops.NewCore(start, end, scale)

	if err != nil {
		return nil, err
	}

	var devops utils.QueryGenerator = &Devops{
		BaseGenerator: g,
		Core:          core,
	}

	return devops, nil
}
