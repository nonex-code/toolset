package securitycode

import (
	"bytes"
	"crypto/md5"
	cryptorand "crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/freetype"
)

// defaultFontData is the built-in typeface, embedded so the package no longer
// depends on the process working directory.
// defaultFontData 是内置字体，通过 embed 打入二进制，避免依赖进程工作目录。
//
//go:embed typeface/ZhiyongElegant.ttf
var defaultFontData []byte

type SimpleSecCode struct {
	SecCodeLength int //验证码长度

	secCode      string           //验证码字符数组
	secCodeKey   string           //验证码key
	characterMap map[int][]string //验证码取值范围 a-z A-Z 0-9
	imgWidth     int              //验证码图片宽度
	imgHeight    int              //验证码图片高度
	pdi          float64
	FontSize     float64
	img          *image.NRGBA
	rander       *rand.Rand
	FontFilePath string //自定义字体路径，为空时使用内置字体
}

func NewSimpleSecCode() *SimpleSecCode {
	s := &SimpleSecCode{
		SecCodeLength: 4,
		FontSize:      32,
		pdi:           95,
		characterMap:  make(map[int][]string),
		rander:        rand.New(rand.NewSource(randomSeed())),
	}
	s.createCharactersMap()
	return s
}

// randomSeed returns an unpredictable seed for the captcha generator.
// randomSeed 为验证码生成器返回一个不可预测的随机种子。
func randomSeed() int64 {
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err == nil {
		return int64(binary.LittleEndian.Uint64(b[:]))
	}
	return time.Now().UnixNano()
}

// Generate creates the captcha text and image. Errors are ignored; use
// GenerateWithError if you need to handle them.
// Generate 生成验证码文本与图片。忽略错误；如需处理错误请使用 GenerateWithError。
func (s *SimpleSecCode) Generate() *SimpleSecCode {
	_, _ = s.GenerateWithError()
	return s
}

// GenerateWithError creates the captcha text and image and reports any error.
// GenerateWithError 生成验证码文本与图片并返回错误。
func (s *SimpleSecCode) GenerateWithError() (*SimpleSecCode, error) {
	s.createSecCodeList()
	if err := s.createImage(); err != nil {
		return s, err
	}
	return s, nil
}

// GetImgBase64 returns the captcha image as a base64-encoded PNG.
// GetImgBase64 返回 base64 编码的 PNG 验证码图片。
func (s *SimpleSecCode) GetImgBase64() (string, error) {
	if s.img == nil {
		return "", errors.New("securitycode: image not generated")
	}
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, s.img); err != nil {
		return "", fmt.Errorf("securitycode: encode png: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// GetSecCode returns the captcha string.
// GetSecCode 获取验证码字符串。
func (s *SimpleSecCode) GetSecCode() string {
	return s.secCode
}

// GetSecCodeKey returns the MD5 key of the captcha string.
// GetSecCodeKey 获取验证码 key。
func (s *SimpleSecCode) GetSecCodeKey() string {
	return s.secCodeKey
}

// CreateImgFile writes the captcha image to path.
// CreateImgFile 将验证码图片写入 path。
func (s *SimpleSecCode) CreateImgFile(path string) error {
	if s.img == nil {
		return errors.New("securitycode: image not generated")
	}
	imgfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer imgfile.Close()
	if err := png.Encode(imgfile, s.img); err != nil {
		return fmt.Errorf("securitycode: encode png: %w", err)
	}
	return nil
}

// 生成验证码字符数组
func (s *SimpleSecCode) createSecCodeList() *SimpleSecCode {
	var build strings.Builder
	for i := 0; i < s.SecCodeLength; i++ {
		j := s.rander.Intn(len(s.characterMap))
		k := s.rander.Intn(len(s.characterMap[j]))
		build.WriteString(s.characterMap[j][k])
	}
	s.secCode = build.String()
	s.secCodeKey = s.key(s.secCode)
	return s
}

// 生成验证码图片
func (s *SimpleSecCode) createImage() error {
	s.imgHeight = int(s.FontSize * 1.2)
	s.imgWidth = int(s.FontSize) * s.SecCodeLength
	img := image.NewNRGBA(image.Rect(0, 0, s.imgWidth, s.imgHeight))

	var rgbRand = []uint8{uint8(s.rander.Intn(85)), uint8(s.rander.Intn(170-85) + 85), uint8(s.rander.Intn(255-170) + 170)}
	// 设置每个点的 RGBA
	for i := 0; i < s.imgWidth*s.imgHeight; i++ {
		img.Set(s.rander.Intn(s.imgWidth), s.rander.Intn(s.imgHeight), color.NRGBA{
			R: rgbRand[s.rander.Intn(3)],
			B: rgbRand[s.rander.Intn(3)],
			G: 245,
			A: 255,
		})
	}

	// Prefer a caller-provided font; otherwise use the embedded default.
	// 优先使用调用方指定的字体，否则使用内置字体。
	fontData := defaultFontData
	if s.FontFilePath != "" {
		if data, err := os.ReadFile(s.FontFilePath); err == nil {
			fontData = data
		}
	}
	font, err := freetype.ParseFont(fontData)
	if err != nil {
		return fmt.Errorf("securitycode: parse font: %w", err)
	}

	f := freetype.NewContext()
	f.SetDPI(s.pdi)
	f.SetFont(font)
	f.SetFontSize(s.FontSize)
	f.SetClip(img.Bounds())
	f.SetDst(img)

	x := s.imgWidth / int(s.FontSize)
	for _, v := range s.secCode {
		f.SetSrc(image.NewUniform(color.RGBA{
			R: rgbRand[s.rander.Intn(3)],
			G: rgbRand[s.rander.Intn(3)],
			B: rgbRand[s.rander.Intn(3)],
			A: 255,
		}))
		pt := freetype.Pt(x, s.imgHeight/2+int(f.PointToFixed(s.FontSize))>>8)
		if _, err := f.DrawString(string(v), pt); err != nil {
			return fmt.Errorf("securitycode: draw string: %w", err)
		}
		x += int(s.FontSize)
	}
	s.img = img
	return nil
}

// 获取验证码对应的key
func (s *SimpleSecCode) key(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 生成验证码取值范围
func (s *SimpleSecCode) createCharactersMap() *SimpleSecCode {
	var smallLetter []string
	var capitalLetter []string
	var number []string

	for i := 97; i < 123; i++ {
		smallLetter = append(smallLetter, string(rune(i)))
	}
	for i := 90; i > 64; i-- {
		capitalLetter = append(capitalLetter, string(rune(i)))
	}
	for i := 0; i < 10; i++ {
		number = append(number, strconv.Itoa(i))
	}
	s.characterMap[0] = smallLetter
	s.characterMap[1] = capitalLetter
	s.characterMap[2] = number
	return s
}

// VerifySecCode compares two captcha strings case-insensitively.
// VerifySecCode 不区分大小写地比较两个验证码字符串。
func VerifySecCode(seccode1, seccode2 string) bool {
	return strings.EqualFold(seccode1, seccode2)
}

// VerifySecCodeCaseSensitive compares two captcha strings exactly (case-sensitive).
// VerifySecCodeCaseSensitive 精确比较两个验证码字符串（区分大小写）。
func VerifySecCodeCaseSensitive(seccode1, seccode2 string) bool {
	return seccode1 == seccode2
}

// VerifySecCodeIgnoreCase is a deprecated alias of VerifySecCode.
//
// Deprecated: the name previously contradicted the behavior. It now matches its
// name (case-insensitive) and is an alias of VerifySecCode. Use VerifySecCode or
// VerifySecCodeCaseSensitive instead.
//
// VerifySecCodeIgnoreCase 是 VerifySecCode 的已废弃别名。现在与名称一致（不区分大小写）。
// 请改用 VerifySecCode 或 VerifySecCodeCaseSensitive。
func VerifySecCodeIgnoreCase(seccode1, seccode2 string) bool {
	return VerifySecCode(seccode1, seccode2)
}
