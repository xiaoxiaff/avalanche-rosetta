package mapper

import "time"

func UnixToUnixMilli(unixSec int64) int64 {
	return time.Unix(unixSec, 0).UnixMilli()
}
