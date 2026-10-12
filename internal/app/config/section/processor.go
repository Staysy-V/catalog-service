package section

type (
	Processor struct {
		WebServer ProcessorWebServer
	}

	ProcessorWebServer struct {
		ListenPort uint32 `split_words:"true"`
	}
)
