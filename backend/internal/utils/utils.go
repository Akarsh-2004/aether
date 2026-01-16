package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID generates a random hexadecimal ID
func GenerateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateSessionID generates a session ID with timestamp
func GenerateSessionID() string {
	timestamp := time.Now().Unix()
	random := GenerateID()
	return fmt.Sprintf("sess_%d_%s", timestamp, random)
}

// Clamp returns value clamped between min and max
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Lerp performs linear interpolation between a and b by t
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// DistanceSquared returns squared distance between two points
func DistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

// Distance returns distance between two points
func Distance(x1, y1, x2, y2 float64) float64 {
	return sqrt(DistanceSquared(x1, y1, x2, y2))
}

// Fast approximation of square root
func sqrt(x float64) float64 {
	// This is a fast approximation using the Quake algorithm
	// For production, consider using math.Sqrt for accuracy
	if x == 0 {
		return 0
	}
	
	// Initial guess
	x2 := x * 0.5
	y := x
	
	// Magic number for IEEE 754 floats
	i := int64(0x5f3759df - (int64(y) >> 1))
	y = float64(i)
	
	// Newton-Raphson iterations
	y = y * (1.5 - (x2 * y * y))
	y = y * (1.5 - (x2 * y * y))
	
	return x * y
}

// Min returns the minimum of two floats
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two floats
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Abs returns absolute value of a float
func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TimeSince returns time duration since given time
func TimeSince(t time.Time) time.Duration {
	return time.Since(t)
}

// FormatDuration formats duration in human readable format
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2f μs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1000000)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}

// SafeStringSlice safely creates a string slice with bounds checking
func SafeStringSlice(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start >= end {
		return ""
	}
	return s[start:end]
}

// Contains checks if slice contains string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Remove removes item from slice
func Remove(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
