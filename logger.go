package golog

import (
	"encoding/json"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"time"
)

var logger *zap.Logger = nil

//获取初始配置数据
func getDefaultJson()*[]byte{
	defaultJson := []byte(`{
    "level":"warn",
    "encoding":"json",
    "outputPaths": ["stdout", "log.txt"],
    "errorOutputPaths": ["stderr"],
    "encoderConfig": {
      "messageKey": "message",
	  "timeKey": "time",
      "levelKey": "level",
      "levelEncoder": "lowercase"
    }
  }`)

	return &defaultJson
}

//时间格式
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

//日志分割
func getLogWriter(path string)*lumberjack.Logger{
	if path == ""{
		return nil
	}

	lj := lumberjack.Logger{
		Filename:   path, // 输出文件
		MaxSize:    5,       // 日志文件最大大小（单位：MB）
		MaxBackups: 3,         // 保留的旧日志文件最大数量
		MaxAge:     30,        // 保存日期
	}
	return &lj
}

//获取Log初始配置
func getConfig()*zap.Config{
	var cfg zap.Config
	er := json.Unmarshal(*getDefaultJson(), &cfg)
	if er != nil{
		panic("log初始化配置出错")
	}
	return &cfg
}


//初始化Log，需传递log配置文件路径，若该配置文件不存在则会以默认初始配置进行初始化
func Init(path string){
	var ler error
	var cfg *zap.Config
	if _, er := os.Stat(path); er != nil{
		cfg = getConfig()
		cfg.EncoderConfig.EncodeTime = timeEncoder
		logger, ler = cfg.Build()
		if ler != nil{
			panic("初始化日志失败，" + ler.Error())
		}

		file, er := os.Create(path)
		if er != nil{
			panic("无法创建配置文件，" + er.Error())
		}
		_, wer := file.Write(*getDefaultJson())
		if wer != nil{
			panic("写入配置文件失败，" + wer.Error())
		}
		_ = file.Close()
	}else{
		file, er := os.Open(path)
		if er != nil{
			panic("无法打开配置文件，" + er.Error())
		}
		data, fer := ioutil.ReadAll(file)
		if fer != nil{
			panic("读取配置文件失败，" + fer.Error())
		}

		cfg = &zap.Config{}
		cer := json.Unmarshal(data, cfg)
		if cer != nil{
			panic("所读取的配置文件存在错误，" + cer.Error())
		}
		cfg.EncoderConfig.EncodeTime = timeEncoder
		logger, ler = cfg.Build()

		if ler != nil{
			panic("初始化日志失败，" + ler.Error())
		}
	}

	zap.ReplaceGlobals(logger)
	out := cfg.OutputPaths
	if len(out) == 2{
		if out[0] == "stdout"{
			zapcore.AddSync(getLogWriter(out[1]))
		}else{
			zapcore.AddSync(getLogWriter(out[0]))
		}
	}else if len(out) == 1 && out[0] != "stdout"{
		zapcore.AddSync(getLogWriter(out[0]))
	}
}

//获取Log单例指针，调用此函数前请先调用初始化函数
func GetLogger() *zap.Logger{
	if logger == nil{
		panic("未调用初始化函数")
	}

	return logger
}
