package image

import (
	"GopherAI/common/image"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalRecognizer *image.ImageRecognizer
	recognizerOnce   sync.Once
	recognizerErr    error
)

// getModelPaths 获取模型文件的绝对路径
// 优先从环境变量读取，否则使用默认的 models/ 目录
func getModelPaths() (modelPath, labelPath string) {
	modelPath = os.Getenv("GOPHERAI_MODEL_PATH")
	labelPath = os.Getenv("GOPHERAI_LABEL_PATH")
	if modelPath != "" && labelPath != "" {
		return
	}
	// 使用可执行文件所在目录下的 models/ 目录
	exe, err := os.Executable()
	if err == nil {
		baseDir := filepath.Dir(exe)
		modelPath = filepath.Join(baseDir, "models", "mobilenetv2-7.onnx")
		labelPath = filepath.Join(baseDir, "models", "imagenet_classes.txt")
		// 检查文件是否存在
		if _, err := os.Stat(modelPath); err == nil {
			return
		}
	}
	// 回退到当前工作目录下的 models/ 目录
	modelPath = "models/mobilenetv2-7.onnx"
	labelPath = "models/imagenet_classes.txt"
	return
}

// getOrCreateRecognizer 获取或创建全局共享的 ImageRecognizer 实例（单例）
func getOrCreateRecognizer() (*image.ImageRecognizer, error) {
	recognizerOnce.Do(func() {
		modelPath, labelPath := getModelPaths()

		// 检查模型文件是否存在
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			recognizerErr = &ModelNotFoundError{
				ModelPath: modelPath,
				Message:   "模型文件不存在，请先下载模型文件放置到 models/ 目录",
			}
			log.Printf("[ImageRecognizer] %v", recognizerErr)
			return
		}
		if _, err := os.Stat(labelPath); os.IsNotExist(err) {
			recognizerErr = &ModelNotFoundError{
				ModelPath: labelPath,
				Message:   "标签文件不存在，请先下载标签文件放置到 models/ 目录",
			}
			log.Printf("[ImageRecognizer] %v", recognizerErr)
			return
		}

		recognizer, err := image.NewImageRecognizer(modelPath, labelPath, 224, 224)
		if err != nil {
			recognizerErr = err
			log.Printf("[ImageRecognizer] NewImageRecognizer fail: %v", err)
			return
		}
		globalRecognizer = recognizer
		log.Println("[ImageRecognizer] Initialized successfully")
	})

	if recognizerErr != nil {
		return nil, recognizerErr
	}

	// 如果 recognizer 已被 Close（理论上不会发生），重置
	if globalRecognizer == nil {
		recognizerOnce = sync.Once{}
		recognizerErr = nil
		return getOrCreateRecognizer()
	}

	return globalRecognizer, nil
}

// ModelNotFoundError 模型未找到错误
type ModelNotFoundError struct {
	ModelPath string
	Message   string
}

func (e *ModelNotFoundError) Error() string {
	return e.Message
}

// IsModelNotFoundError 判断是否为模型未找到错误
func IsModelNotFoundError(err error) bool {
	_, ok := err.(*ModelNotFoundError)
	return ok
}

// RecognizeImage 识别图片
func RecognizeImage(file *multipart.FileHeader) (string, error) {
	// 获取全局共享的识别器（只会创建一次）
	recognizer, err := getOrCreateRecognizer()
	if err != nil {
		log.Println("RecognizeImage getOrCreateRecognizer fail: ", err)
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		log.Println("file open fail err is : ", err)
		return "", err
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		log.Println("io.ReadAll fail err is : ", err)
		return "", err
	}

	return recognizer.PredictFromBuffer(buf)
}

// Close 关闭全局识别器（通常在程序退出时调用）
func Close() {
	if globalRecognizer != nil {
		globalRecognizer.Close()
		globalRecognizer = nil
	}
}
