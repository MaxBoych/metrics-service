package values

import "time"

// Мапа необходима согласно ТЗ инкремента 13:
// Обработка Retriable-ошибок.
// Количество повторов должно быть ограничено тремя дополнительными попытками.
// Интервалы между повторами должны увеличиваться: 1s, 3s, 5s.
var RetryIntervals = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}
