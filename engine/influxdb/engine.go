package influxdb

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/influxdb/influxdb/client"
	"github.com/nvcook42/morgoth/engine"
	metric "github.com/nvcook42/morgoth/metric/types"
	"github.com/nvcook42/morgoth/schedule"
	"math"
	"time"
)

const (
	metricPrefix                    = "m."
	metricPrefixPtrn metric.Pattern = "^m\\."
)

type InfluxDBEngine struct {
	config *InfluxDBConf
	client *client.Client
}

func (self *InfluxDBEngine) Initialize() error {
	client, err := connect(self.config)
	if err != nil {
		return err
	}
	self.client = client
	return nil
}

func (self *InfluxDBEngine) GetReader() engine.Reader {
	return self
}

func (self *InfluxDBEngine) GetWriter() engine.Writer {
	return self
}

func (self *InfluxDBEngine) ConfigureSchedule(schedule schedule.Schedule) error {

	result, err := self.client.Query("list continuous queries")
	if err != nil {
		return err
	}

	existing := make([]string, 0)
	for _, series := range result {
		for _, row := range series.GetPoints() {
			existing = append(existing, row[2].(string))
		}
	}

	for _, rotation := range schedule.Rotations {
		resolution := int64(math.Ceil(rotation.Resolution.Seconds()))
		q := fmt.Sprintf("select first(value) as value from /^m\\..*/ group by time(%ds) into %s.:series_name", resolution, rotation.String())
		found := false
		for _, e := range existing {
			if e == q {
				found = true
				break
			}
		}
		if found {
			continue
		}
		log.Infof("Creating continuous query '%s'", q)
		_, err = self.client.Query(q)
		if err != nil {
			return err
		}
	}

	return nil

}

//////////////////////
// Writer Methods
//////////////////////

func (self *InfluxDBEngine) Insert(datetime time.Time, metric metric.MetricID, value float64) {
	series := new(client.Series)
	series.Name = metricPrefix + string(metric)
	series.Columns = []string{
		"time",
		"value",
	}
	series.Points = [][]interface{}{
		[]interface{}{datetime.Unix(), value},
	}
	self.client.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second)
}

func (self *InfluxDBEngine) RecordAnomalous(metric metric.MetricID, start, stop time.Time) {
}

func (self *InfluxDBEngine) DeleteMetric(metric metric.MetricID) {
}

//////////////////////
// Reader Methods
//////////////////////

func (self *InfluxDBEngine) GetMetrics() []metric.MetricID {
	return nil
}
func (self *InfluxDBEngine) GetData(metric metric.MetricID, start, stop time.Time, step time.Duration) []engine.Point {
	result, err := self.client.Query(
		fmt.Sprintf("select value from rot.1.60.m.%s where time > %ds and time < %ds",
			metric,
			start.Unix(),
			stop.Unix(),
		),
		client.Second,
	)

	if err != nil {
		log.Error(err.Error())
		return []engine.Point{}
	}
	if len(result) == 0 {
		return []engine.Point{}
	}

	series := result[0]
	points := series.GetPoints()
	log.Debug(len(points))
	data := make([]engine.Point, len(points))
	for i, row := range points {
		sec := int64(row[0].(float64))
		data[i].Time = time.Unix(sec, 0)
		data[i].Value = row[2].(float64)
	}
	return data
}

func (self *InfluxDBEngine) GetAnomalies(metric metric.MetricID, start, stop time.Time) []engine.Anomaly {
	return nil
}
func (self *InfluxDBEngine) GetHistogram(metric metric.MetricID, nbins uint, start, stop time.Time) *engine.Histogram {
	hist := new(engine.Histogram)

	m_min := -2.0
	step_size := 1.0

	result, err := self.client.Query(fmt.Sprintf("select count(value), histogram(value, 1, -2, 2) from rot.10.300.%s where time > %ds and time < %ds",
		metricPrefix+metric,
		start.Unix(),
		stop.Unix(),
	))

	if err != nil {
		log.Error(err.Error())
		return hist
	}

	series := result[0]
	points := series.GetPoints()
	hist.Bins = make([]float64, len(points))
	for _, row := range points {
		total := row[1].(float64)
		bucketStart := row[2].(float64)
		count := row[3].(float64)
		i := int((bucketStart - m_min) / step_size)
		hist.Bins[i] = count / total
		hist.Count = uint(total)
	}

	return hist

}
func (self *InfluxDBEngine) GetPercentile(metric metric.MetricID, percentile float64, start, stop time.Time) float64 {
	return 0.0
}
