package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// ToFxLogger создает fxevent.Logger из кастомного логгера
func (l *Logger) ToFxLogger() fxevent.Logger {
	return &fxLogger{logger: l}
}

// FxLogger адаптер для интеграции с fx.WithLogger
type fxLogger struct {
	logger *Logger
}

// LogEvent реализует интерфейс fxevent.Logger
func (f *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		f.logger.Info("fx: OnStart hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName))

	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			f.logger.Error("fx: OnStart hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Duration("runtime", e.Runtime),
				zap.Error(e.Err))
		} else {
			f.logger.Info("fx: OnStart hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Duration("runtime", e.Runtime))
		}

	case *fxevent.OnStopExecuting:
		f.logger.Info("fx: OnStop hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName))

	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			f.logger.Error("fx: OnStop hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Duration("runtime", e.Runtime),
				zap.Error(e.Err))
		} else {
			f.logger.Info("fx: OnStop hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Duration("runtime", e.Runtime))
		}

	case *fxevent.Provided:
		if e.Err != nil {
			f.logger.Error("fx: provided constructor failed",
				zap.String("constructor", e.ConstructorName),
				zap.Strings("output_types", e.OutputTypeNames),
				zap.String("module", e.ModuleName),
				zap.Error(e.Err))
		} else {
			f.logger.Debug("fx: provided constructor",
				zap.String("constructor", e.ConstructorName),
				zap.Strings("output_types", e.OutputTypeNames),
				zap.String("module", e.ModuleName))
		}

	case *fxevent.Invoked:
		if e.Err != nil {
			f.logger.Error("fx: invoke failed",
				zap.String("function", e.FunctionName),
				zap.String("module", e.ModuleName),
				zap.Error(e.Err))
		} else {
			f.logger.Debug("fx: invoked",
				zap.String("function", e.FunctionName),
				zap.String("module", e.ModuleName))
		}

	case *fxevent.Invoking:
		f.logger.Debug("fx: invoking",
			zap.String("function", e.FunctionName),
			zap.String("module", e.ModuleName))

	case *fxevent.Started:
		if e.Err != nil {
			f.logger.Error("fx: start failed", zap.Error(e.Err))
		} else {
			f.logger.Info("fx: started")
		}

	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			f.logger.Error("fx: custom logger initialization failed",
				zap.String("constructor", e.ConstructorName),
				zap.Error(e.Err))
		} else {
			f.logger.Debug("fx: initialized custom logger",
				zap.String("constructor", e.ConstructorName))
		}

	default:
		f.logger.Debug("fx: unhandled event", zap.Any("event", e))
	}
}
