package consumer

import (
	"music-library/config"
	"music-library/services"
)

func StartConsumer() {
	config.LoadConfig()
	services.ConsumeSongQueue()
}
