// tsbs_run_queries_redistimeseries speed tests RedisTimeSeries using requests from stdin or file
//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	redistimeseries "github.com/RedisTimeSeries/redistimeseries-go"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/timescale/tsbs/internal/utils"
	"github.com/timescale/tsbs/query"
	"log"
	"sort"
	"strings"
	"time"
)

// Program option vars:
var (
	host        string
	showExplain bool
	//	scale        uint64
)

// Global vars:
var (
	runner        *query.BenchmarkRunner
	cmdMrange     = []byte("TS.MRANGE")
	cmdQueryIndex = []byte("TS.QUERYINDEX")
)

var (
	redisConnector *redistimeseries.Client
)

// Parse args:
func init() {
	var config query.BenchmarkRunnerConfig
	config.AddToFlagSet(pflag.CommandLine)

	pflag.StringVar(&host, "host", "localhost:6379", "Redis host address and port")

	pflag.Parse()

	err := utils.SetupConfigFile()

	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("unable to decode config: %s", err))
	}
	runner = query.NewBenchmarkRunner(config)

	redisConnector = redistimeseries.NewClient(
		host, runner.DatabaseName(), nil)
}

func main() {
	runner.Run(&query.RedisTimeSeriesPool, newProcessor)
}

type queryExecutorOptions struct {
	showExplain   bool
	debug         bool
	printResponse bool
}

type processor struct {
	opts *queryExecutorOptions
}

func newProcessor() query.Processor { return &processor{} }

func (p *processor) Init(numWorker int) {
	p.opts = &queryExecutorOptions{
		showExplain:   showExplain,
		debug:         runner.DebugLevel() > 0,
		printResponse: runner.DoPrintResponses(),
	}
}

func mapRows(r *sql.Rows) []map[string]interface{} {
	rows := []map[string]interface{}{}
	cols, _ := r.Columns()
	for r.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(cols))
		for i := range values {
			values[i] = new(interface{})
		}

		err := r.Scan(values...)
		if err != nil {
			panic(errors.Wrap(err, "error while reading values"))
		}

		for i, column := range cols {
			row[column] = *values[i].(*interface{})
		}
		rows = append(rows, row)
	}
	return rows
}

// prettyPrintResponseRange prints a Query and its response in JSON format with two
// keys: 'query' which has a value of the RedisTimeseries query used to generate the second key
// 'results' which is an array of each element in the return set.
func prettyPrintResponseRange(responses []interface{}, q *query.RedisTimeSeries) {
	full := make(map[string]interface{})
	for idx, qry := range q.RedisQueries {
		resp := make(map[string]interface{})
		fullcmd := append([][]byte{q.CommandNames[idx]}, qry...)
		resp["query"] = strings.Join(ByteArrayToStringArray(fullcmd), " ")

		res := responses[idx]
		switch v := res.(type) {
		case []redistimeseries.Range:
			resp["client_side_work"] = q.ReduceSeries || q.SingleGroupByTimestamp
			rows := []map[string]interface{}{}
			for _, r := range res.([]redistimeseries.Range) {
				row := make(map[string]interface{})
				values := make(map[string]interface{})
				values["datapoints"] = r.DataPoints
				values["labels"] = r.Labels
				row[r.Name] = values
				rows = append(rows, row)
			}
			resp["results"] = rows
		case redistimeseries.Range:
			resp["client_side_work"] = q.ReduceSeries || q.SingleGroupByTimestamp
			resp["results"] = res.(redistimeseries.Range)
		case query.MultiRange:
			resp["client_side_work"] = q.ReduceSeries || q.SingleGroupByTimestamp
			query_result := map[string]interface{}{}
			converted := res.(query.MultiRange)
			query_result["names"] = converted.Names
			query_result["labels"] = converted.Labels
			datapoints := make([]query.MultiDataPoint, 0, len(converted.DataPoints))
			var keys []int
			for k := range converted.DataPoints {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			for _, k := range keys {
				dp := converted.DataPoints[int64(k)]
				time_str := time.Unix(dp.Timestamp/1000, 0).Format(time.RFC3339)
				dp.HumanReadbleTime = &time_str
				datapoints = append(datapoints, dp)
			}
			query_result["datapoints"] = datapoints
			resp["results"] = query_result
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}

		full[fmt.Sprintf("query %d", idx+1)] = resp
	}

	line, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(line) + "\n")
}

func (p *processor) ProcessQuery(q query.Query, isWarm bool) (queryStats []*query.Stat, err error) {

	// No need to run again for EXPLAIN
	if isWarm && p.opts.showExplain {
		return nil, nil
	}
	tq := q.(*query.RedisTimeSeries)
	var parsedResponses = make([]interface{}, 0, 0)

	var cmds = make([][]interface{}, 0, 0)
	for _, qry := range tq.RedisQueries {
		cmds = append(cmds, ByteArrayToInterfaceArray(qry))
	}
	conn := redisConnector.Pool.Get()

	start := time.Now()
	for idx, commandArgs := range cmds {
		var result interface{}
		if p.opts.debug {
			fmt.Println(fmt.Sprintf("Issuing command (%s %s)", string(tq.CommandNames[idx]), strings.Join(ByteArrayToStringArray(tq.RedisQueries[idx]), " ")))
		}
		res, err := conn.Do(string(tq.CommandNames[idx]), commandArgs...)
		if err != nil {
			log.Fatalf("Command (%s %s) failed with error: %v\n", string(tq.CommandNames[idx]), strings.Join(ByteArrayToStringArray(tq.RedisQueries[idx]), " "), err)
		}
		if err != nil {
			return nil, err
		}
		if bytes.Compare(tq.CommandNames[idx], cmdMrange) == 0 {
			parsedRes, err := redistimeseries.ParseRanges(res)
			result = parsedRes

			if err != nil {
				return nil, err
			}
			if tq.SingleGroupByTimestamp {
				result = query.MergeSeriesOnTimestamp(parsedRes)
			}
			if tq.ReduceSeries {
				result, err = query.ReduceSeriesOnTimestampBy(parsedRes, query.MaxReducerSeriesDatapoints )
				if err != nil {
					return nil, err
				}
			}

		} else if bytes.Compare(tq.CommandNames[idx], cmdQueryIndex) == 0 {
			var parsedRes = make([]redistimeseries.Range, 0, 0)
			parsedResponses = append(parsedResponses, parsedRes)
		}
		parsedResponses = append(parsedResponses, result)
	}
	took := float64(time.Since(start).Nanoseconds()) / 1e6
	if p.opts.printResponse {
		prettyPrintResponseRange(parsedResponses, tq)
	}
	stat := query.GetStat()
	stat.Init(q.HumanLabelName(), took)
	queryStats = []*query.Stat{stat}

	return queryStats, err
}

func ByteArrayToInterfaceArray(qry [][]byte) []interface{} {
	commandArgs := make([]interface{}, len(qry))
	for i := 0; i < len(qry); i++ {
		commandArgs[i] = qry[i]
	}
	return commandArgs
}

func ByteArrayToStringArray(qry [][]byte) []string {
	commandArgs := make([]string, len(qry))
	for i := 0; i < len(qry); i++ {
		commandArgs[i] = string(qry[i])
	}
	return commandArgs
}
