package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"math/rand"
	"strings"
)

// MakeDir make directory path is not exist.
func GetDir(dirPath string) string {
	if !strings.HasPrefix(dirPath, "/") && !strings.HasPrefix(dirPath, "./") {
		dirPath = "./" + dirPath
	}
	if !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}
	/*
		s, err := os.Stat(dirPath)
		if err != nil {
			err = os.Mkdir(dirPath, 0750)
			if err != nil {
				return err
			}
		} else {
			if !s.IsDir() {
				return nil
			}
		}
	*/
	return dirPath
}

// B64Decode  decode base64 string
func B64Decode(d string) string {
	b, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		return ""
	}
	return string(b)
}

// B64Encode encode string with base64
func B64Encode(d string) string {
	return base64.StdEncoding.EncodeToString([]byte(d))
}

// ConvertOctetStreamToJPEG 将application/octet-stream数据转换为image/jpeg格式
func ConvertOctetStreamToJPEG(data []byte) ([]byte, error) {
	// 将数据解码为image.Image
	img, err := decodeOctetStream(data)
	if err != nil {
		return nil, err
	}

	// 将image.Image编码为JPEG格式
	return encodeJPEG(img)
}

// decodeOctetStream 将application/octet-stream数据解码为image.Image
func decodeOctetStream(data []byte) (image.Image, error) {
	// 在这里执行你的解码逻辑
	// 这里假设你使用的是base64编码，你可以根据实际情况进行修改

	// 先将数据进行base64解码
	decodedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	// 将解码后的数据读取为image.Image
	img, _, err := image.Decode(bytes.NewReader(decodedData))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// encodeJPEG 将image.Image编码为JPEG格式
func encodeJPEG(img image.Image) ([]byte, error) {
	// 创建一个字节缓冲区
	buf := new(bytes.Buffer)

	// 将image.Image编码为JPEG格式并将结果写入缓冲区
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		return nil, err
	}

	// 将缓冲区中的数据读取为字节数组
	data, err := io.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/*
func main() {
	// 示例用法
	octetStreamData := []byte("application/octet-stream data")
	jpegData, err := ConvertOctetStreamToJPEG(octetStreamData)
	if err != nil {
		// 处理错误
		return
	}

	// 使用jpegData进行后续操作，比如保存为文件等
}
*/

// DetectImageType 识别图片类型
func DetectImageType(data []byte) (string, error) {
	// 将数据的前几个字节作为文件头进行判断
	fileHeader := data[:4]

	// 根据文件头判断图片类型
	switch {
	case isJPEG(fileHeader):
		return "image/jpeg", nil
	case isPNG(fileHeader):
		return "image/png", nil
	case isGIF(fileHeader):
		return "image/gif", nil
	default:
		return "", errors.New("unknown image type")
	}
}

// 判断是否为JPEG格式
func isJPEG(fileHeader []byte) bool {
	return len(fileHeader) >= 2 &&
		fileHeader[0] == 0xff && fileHeader[1] == 0xd8
}

// 判断是否为PNG格式
func isPNG(fileHeader []byte) bool {
	return len(fileHeader) >= 8 &&
		fileHeader[0] == 0x89 &&
		fileHeader[1] == 0x50 &&
		fileHeader[2] == 0x4e &&
		fileHeader[3] == 0x47 &&
		fileHeader[4] == 0x0d &&
		fileHeader[5] == 0x0a &&
		fileHeader[6] == 0x1a &&
		fileHeader[7] == 0x0a
}

// 判断是否为GIF格式
func isGIF(fileHeader []byte) bool {
	return len(fileHeader) >= 6 &&
		((fileHeader[0] == 0x47 && fileHeader[1] == 0x49 && fileHeader[2] == 0x46 && fileHeader[3] == 0x38 && (fileHeader[4] == 0x39 || fileHeader[4] == 0x37) && fileHeader[5] == 0x61) ||
			(fileHeader[0] == 0x47 && fileHeader[1] == 0x49 && fileHeader[2] == 0x46 && fileHeader[3] == 0x38 && fileHeader[4] == 0x39 && fileHeader[5] == 0x61))
}
func RandomStr(n int) string {
	var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = defaultLetters[rand.Intn(len(defaultLetters))]
	}

	return string(b)
}

func GetFileSubfix(mimtype string) string {
	var MimeExtensionSanityOverrides = map[string]string{
		"application/x-abiword":        ".abw",
		"application/x-freearc":        ".arc",
		"video/x-msvideo":              ".avi",
		"application/vnd.amazon.ebook": ".azw",
		"application/octet-stream":     ".bin",
		"image/bmp":                    ".bmp",
		"application/x-bzip":           ".bz",
		"application/x-bzip2":          ".bz2",
		"application/x-csh":            ".csh",
		"text/css":                     ".css",
		"text/csv":                     ".csv",
		"application/msword":           ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-fontobject":                                           ".eot",
		"application/epub+zip":                                                    ".epub",
		"image/gif":                                                               ".gif",
		"text/html":                                                               ".html",
		"image/vnd.microsoft.icon":                                                ".ico",
		"text/calendar":                                                           ".ics",
		"application/java-archive":                                                ".jar",
		".jpg":                                                                    ".jpeg",
		"text/javascript":                                                         ".js",
		"application/json":                                                        ".json",
		"application/ld+json":                                                     ".jsonld",
		"audio/midi":                                                              ".mid",
		"audio/x-midi":                                                            ".midi",
		"audio/mpeg":                                                              ".mp3",
		"video/mpeg":                                                              ".mpeg",
		"application/vnd.apple.installer+xml":                                     ".mpkg",
		"application/vnd.oasis.opendocument.presentation": ".odp",
		"application/vnd.oasis.opendocument.spreadsheet":  ".ods",
		"application/vnd.oasis.opendocument.text":         ".odt",
		"audio/ogg":                     ".oga",
		"video/ogg":                     ".ogv",
		"application/ogg":               ".ogx",
		"font/otf":                      ".otf",
		"image/png":                     ".png",
		"application/pdf":               ".pdf",
		"application/vnd.ms-powerpoint": ".ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"application/x-rar-compressed":                                              ".rar",
		"application/rtf":                                                           ".rtf",
		"application/x-sh":                                                          ".sh",
		"image/svg+xml":                                                             ".svg",
		"application/x-shockwave-flash":                                             ".swf",
		"application/x-tar":                                                         ".tar",
		"image/tiff":                                                                ".tif .tiff",
		"font/ttf":                                                                  ".ttf",
		"text/plain":                                                                ".txt",
		"application/vnd.visio":                                                     ".vsd",
		"audio/wav":                                                                 ".wav",
		"audio/webm":                                                                ".weba",
		"video/webm":                                                                ".webm",
		"image/webp":                                                                ".webp",
		"font/woff":                                                                 ".woff",
		"font/woff2":                                                                ".woff2",
		"application/xhtml+xml":                                                     ".xhtml",
		"application/vnd.ms-excel":                                                  ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
		"XML":                             ".xml",
		"application/vnd.mozilla.xul+xml": ".xul",
		"application/zip":                 ".zip",
		"video/3gpp":                      ".3gp",
		"video/3gpp2":                     ".3g2",
		"application/x-7z-compressed":     ".7z",
	}
	return MimeExtensionSanityOverrides[mimtype]
}
