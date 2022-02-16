package business

type Logger struct {
  forecastProvider ForecastProvider
  observers []Observer
}

func (l Logger) Execute(c Command) {
  if c.IsFecthForecasts() {
    fcs := l.forecastProvider.Get(SearchRequest{})

    for _, observer := range l.observers {
      observer.Notify(fcs)
    }
  }
}

func MakeLogger(fcp ForecastProvider, obs []Observer) Logger {
  return Logger{forecastProvider: fcp, observers: obs}
}
