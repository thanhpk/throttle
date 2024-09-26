package throttle

import (
	"hash/crc32"
	"sync"
	"time"
)

const RATELIMITSHARD = 128

var _dataPerMinS [RATELIMITSHARD]map[int64]map[string]int64  // min -> key -> count
var _dataPerHourS [RATELIMITSHARD]map[int64]map[string]int64 // min -> key -> count
var _ratelimitLockS [RATELIMITSHARD]*sync.Mutex

func init() {
	for i := 0; i < RATELIMITSHARD; i++ {
		_dataPerMinS[i] = map[int64]map[string]int64{}
		_dataPerHourS[i] = map[int64]map[string]int64{}
		_ratelimitLockS[i] = &sync.Mutex{}
	}

	// clean
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			for shard := range RATELIMITSHARD {
				_ratelimitLock := _ratelimitLockS[shard]
				_ratelimitLock.Lock()
				nowMin := time.Now().Unix() / 60
				nowHour := nowMin / 60
				copyDataPerMin := map[int64]map[string]int64{}
				for min := nowMin - 5; min < nowMin+2; min++ {
					copyDataPerMin[min] = _dataPerMinS[shard][min]
				}
				_dataPerMinS[shard] = copyDataPerMin

				copyDataPerHour := map[int64]map[string]int64{}
				for hour := nowHour - 2; hour < nowHour+2; hour++ {
					copyDataPerHour[hour] = _dataPerHourS[shard][hour]
				}
				_dataPerHourS[shard] = copyDataPerHour
				_ratelimitLock.Unlock()
			}
		}
	}()
}

// req per min
func RateLimit(key string, rpm, rph int64) bool {
	nowMin := time.Now().Unix() / 60
	nowHour := nowMin / 60
	shard := crc32.ChecksumIEEE([]byte(key)) % RATELIMITSHARD

	_ratelimitLockS[shard].Lock()
	defer _ratelimitLockS[shard].Unlock()

	if _dataPerMinS[shard][nowMin] == nil {
		_dataPerMinS[shard][nowMin] = map[string]int64{}
	}
	perMin := _dataPerMinS[shard][nowMin]
	perMin[key] = perMin[key] + 1
	failed := false
	if perMin[key] > rpm {
		failed = true
	}

	if _dataPerHourS[shard][nowHour] == nil {
		_dataPerHourS[shard][nowHour] = map[string]int64{}
	}
	perHour := _dataPerHourS[shard][nowHour]
	perHour[key] = perHour[key] + 1
	return failed || perHour[key] > rph
}
