package throttle

import (
	"hash/fnv"
	"sync"
	"time"
)

const RATELIMITSHARD = 128

var _dataPerMinS [RATELIMITSHARD]map[int64]map[string]int64  // min -> key -> count
var _dataPerHourS [RATELIMITSHARD]map[int64]map[string]int64 // hour -> key -> count
var _dataPerDayS [RATELIMITSHARD]map[int64]map[string]int64  // day -> key -> count
var _ratelimitLockS [RATELIMITSHARD]*sync.Mutex

func init() {
	for i := 0; i < RATELIMITSHARD; i++ {
		_dataPerMinS[i] = map[int64]map[string]int64{}
		_dataPerHourS[i] = map[int64]map[string]int64{}
		_dataPerDayS[i] = map[int64]map[string]int64{}
		_ratelimitLockS[i] = &sync.Mutex{}
	}

	// clean
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			for shard := 0; shard < RATELIMITSHARD; shard++ {
				_ratelimitLock := _ratelimitLockS[shard]
				_ratelimitLock.Lock()
				nowMin := time.Now().Unix() / 60
				copyDataPerMin := map[int64]map[string]int64{}
				for min := nowMin - 5; min < nowMin+2; min++ {
					copyDataPerMin[min] = _dataPerMinS[shard][min]
				}
				_dataPerMinS[shard] = copyDataPerMin

				// clean hour
				nowHour := nowMin / 60
				copyDataPerHour := map[int64]map[string]int64{}
				for hour := nowHour - 2; hour < nowHour+2; hour++ {
					copyDataPerHour[hour] = _dataPerHourS[shard][hour]
				}
				_dataPerHourS[shard] = copyDataPerHour

				// clean day
				nowDay := nowMin / 1440
				copyDataPerDay := map[int64]map[string]int64{}
				for day := nowDay - 2; day < nowDay+2; day++ {
					copyDataPerDay[day] = _dataPerDayS[shard][day]
				}
				_dataPerDayS[shard] = copyDataPerDay
				_ratelimitLock.Unlock()
			}
		}
	}()
}

// RateLimit checks if a key has exceeded a rate limit.
// It returns true if the request should be rejected.
func RateLimit(key string, rpm, rph, rpd int64) bool {
	nowMin := time.Now().Unix() / 60
	nowHour := nowMin / 60
	nowDay := nowMin / 1440
	h := fnv.New32a()
	h.Write([]byte(key))
	shard := h.Sum32() % RATELIMITSHARD

	_ratelimitLockS[shard].Lock()
	defer _ratelimitLockS[shard].Unlock()

	if rpm > 0 {
		if _dataPerMinS[shard][nowMin] == nil {
			_dataPerMinS[shard][nowMin] = map[string]int64{}
		}
		perMin := _dataPerMinS[shard][nowMin]
		perMin[key] = perMin[key] + 1
		if perMin[key] > rpm {
			return true
		}
	}

	if rph > 0 {
		if _dataPerHourS[shard][nowHour] == nil {
			_dataPerHourS[shard][nowHour] = map[string]int64{}
		}
		perHour := _dataPerHourS[shard][nowHour]
		perHour[key] = perHour[key] + 1
		if perHour[key] > rph {
			return true
		}
	}

	if rpd > 0 {
		if _dataPerDayS[shard][nowDay] == nil {
			_dataPerDayS[shard][nowDay] = map[string]int64{}
		}
		perDay := _dataPerDayS[shard][nowDay]
		perDay[key] = perDay[key] + 1
		if perDay[key] > rpd {
			return true
		}
	}
	return false
}
