package cron

import (
	"encoding/json"
	"errors"
	"fmt"
	// "log"
	"strconv"
	"strings"
	"time"
)

var (
	localZone, _        = time.LoadLocation("Local")
	defaultFirstTime, _ = time.ParseInLocation("20060102150405", "19700101000000", localZone)
)

type CronSetting struct {
	FirstTimeStr     string        `json:"firstTimeStr,omitempty"`
	Interval         string        `json:"interval"` // 1h | 100d
	IntervalDuration time.Duration `json:"intervalDuration"`
	// Repeat     bool     `json:"repeat"`

	ClockLimitStart string `json:"clockLimitStart,omitempty"`
	ClockLimitEnd   string `json:"clockLimitEnd,omitempty"`
	WeekLimit       string `json:"weekLimit,omitempty"` //

	// nextRuntime time.Time
	firstTime time.Time
}

// func (c *CronSetting) UnmarshalJSON(data []byte) error {
// 	return newSetting(data, c)
// }

func newSetting(setting_bytes []byte, s *CronSetting) error {
	err := json.Unmarshal(setting_bytes, s)
	if err != nil {
		return err
	}

	return s.Init()
}

func NewSetting(setting_bytes []byte) (s *CronSetting, err error) {
	s = new(CronSetting)
	err = newSetting(setting_bytes, s)
	return
}

func (s *CronSetting) Init() (err error) {
	// interval
	if s.Interval == "" {
		return errors.New("interval is required")
	}

	// intervalDuration
	if s.IntervalDuration.Seconds() == 0 {
		s.IntervalDuration, err = s.getIntervalDuration()
		if err != nil {
			return
		}
	}

	// firstTimeStr
	if s.FirstTimeStr != "" {
		// log.Printf("get firsttime %s", s.FirstTimeStr)
		s.firstTime, err = time.ParseInLocation("20060102150405", s.FirstTimeStr, localZone)
		if err != nil {
			return
		}

		// log.Printf("get firsttime %s", s.firstTime.String())
	} else {
		s.firstTime = defaultFirstTime
	}

	return
}

func (s *CronSetting) getIntervalDuration() (time.Duration, error) {
	var duration_desc = s.Interval
	if strings.HasSuffix(duration_desc, "d") {
		days, err := strconv.Atoi(duration_desc[:(len(duration_desc) - 1)])
		if err != nil {
			return 0, err
		}
		return time.ParseDuration(fmt.Sprintf("%dh", days*24))
	} else {
		return time.ParseDuration(duration_desc)
	}
}

func (s *CronSetting) NextRunTime(curtime time.Time) (next_run time.Time) {
	var duration = s.IntervalDuration
	if s.WeekLimit == "weekday" {
		if curtime.Weekday() == time.Sunday {
			curtime = curtime.Truncate(24 * time.Hour).Add(16*time.Hour - time.Second)
		} else if curtime.Weekday() == time.Saturday {
			curtime = curtime.Truncate(24 * time.Hour).Add(40*time.Hour - time.Second)
		}
	}

	if s.firstTime.After(curtime) {
		return s.firstTime
	}

	// log.Printf("duration:%s", duration.String())
	// log.Printf("firstTime:%s", s.firstTime.String())
	// log.Printf("curtime:%s", curtime.String())

	var loopCouont = curtime.Sub(s.firstTime) / duration
	// log.Printf("loopCouont:%d", loopCouont)
	next_run = s.firstTime.Add(duration * (loopCouont + 1))

	if len(s.ClockLimitStart) == 5 {
		if strings.Compare(next_run.Format("15:04"), s.ClockLimitStart) < 0 {
			hour, _ := strconv.Atoi(s.ClockLimitStart[:2])
			minute, _ := strconv.Atoi(s.ClockLimitStart[3:5])
			return s.NextRunTime(time.Date(
				next_run.Year(), next_run.Month(), next_run.Day(),
				hour, minute, next_run.Second(),
				next_run.Nanosecond(), next_run.Location()).
				Add(0 - time.Second),
			)
		}
	}
	if len(s.ClockLimitEnd) == 5 {
		if strings.Compare(next_run.Format("15:04"), s.ClockLimitEnd) > 0 {
			return s.NextRunTime(
				time.Date(
					next_run.Year(), next_run.Month(), next_run.Day(),
					23, 59, 59, 0, next_run.Location(),
				),
			)
		}
	}

	return
}

func (s *CronSetting) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
