package securitycode

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	FontFilePath string
}

func NewSimpleSecCode() *SimpleSecCode {
	s := &SimpleSecCode{
		SecCodeLength: 4,
		FontSize:      32,
		pdi:           95,
		characterMap:  make(map[int][]string),
		rander:        rand.New(rand.NewSource(time.Now().UnixNano())),
		FontFilePath:  "./securitycode/typeface/ZhiyongElegant.ttf",
	}
	s.createCharactersMap()
	return s
}

// Generate 生成验证码
func (s *SimpleSecCode) Generate() *SimpleSecCode {
	s.createSecCodeList()
	s.createImage()
	return s
}

// GetImgBase64 获取验证码图片 base64
func (s *SimpleSecCode) GetImgBase64() (string, error) {
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, s.img)
	if err != nil {
		log.Println(err)
		return "", errors.New(err.Error())
	}
	sc := base64.StdEncoding.EncodeToString(buffer.Bytes())
	return sc, nil
}

// GetSecCode 获取验证码字符串
func (s *SimpleSecCode) GetSecCode() string {
	return s.secCode
}

// GetSecCodeKey 获取验证码key
func (s *SimpleSecCode) GetSecCodeKey() string {
	return s.secCodeKey
}

// CreateImgFile 创建图片文件 ./test.png
func (s *SimpleSecCode) CreateImgFile(path string) error {
	imgfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer imgfile.Close()
	err = png.Encode(imgfile, s.img)
	if err != nil {
		return err
	}
	return nil
}

// 生成验证码字符数组
func (s *SimpleSecCode) createSecCodeList() *SimpleSecCode {
	chatMap := s.createCharactersMap().characterMap
	var build strings.Builder
	for i := 0; i < s.SecCodeLength; i++ {
		j := s.rander.Intn(len(chatMap))
		k := s.rander.Intn(len(chatMap[j]))
		chat := chatMap[j][k]
		build.WriteString(chat)
	}
	secCode := build.String()
	s.secCode = secCode
	s.secCodeKey = s.key(secCode)
	return s
}

// 生成验证码图片
func (s *SimpleSecCode) createImage() {
	s.imgHeight = int(s.FontSize * 1.2)
	s.imgWidth = int(s.FontSize) * s.SecCodeLength
	img := image.NewNRGBA(image.Rect(0, 0, s.imgWidth, s.imgHeight))
	var rgbRand = []uint8{uint8(s.rander.Intn(85)), uint8(s.rander.Intn(170-85) + 85), uint8(s.rander.Intn(255-170) + 170)}
	//设置每个点的 RGBA (Red,Green,Blue,Alpha(设置透明度))
	for i := 0; i < (s.imgWidth*s.imgHeight)/1; i++ {

		img.Set(rand.Intn(s.imgWidth), s.rander.Intn(s.imgHeight), color.NRGBA{
			R: rgbRand[s.rander.Intn(3)],
			//G: rgbRand[rander.Intn(3)],
			B: rgbRand[s.rander.Intn(3)],
			//R: 245,
			G: 245,
			//B: 220,
			//设定alpha图片的透明度 0透明1不透明
			A: 255,
		})
	}
	//读取字体数据
	//fontBytes, err := ioutil.ReadFile("./Kurland.ttf")
	fontBytes, err := ioutil.ReadFile(s.FontFilePath)

	if err != nil {
		log.Println(err)
	}
	//载入字体数据
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println("load front fail", err)
	}
	f := freetype.NewContext()
	//设置分辨率
	f.SetDPI(s.pdi)
	//设置字体
	f.SetFont(font)
	//设置尺寸
	f.SetFontSize(26)
	f.SetClip(img.Bounds())
	//设置输出的图片
	f.SetDst(img)

	x := s.imgWidth / int(s.FontSize)
	for _, v := range s.secCode {
		//设置字体颜色
		f.SetSrc(image.NewUniform(color.RGBA{R: rgbRand[s.rander.Intn(3)], G: rgbRand[s.rander.Intn(3)], B: rgbRand[s.rander.Intn(3)], A: 255}))

		//设置字体的位置
		pt := freetype.Pt(x, s.imgHeight/2+int(f.PointToFixed(26))>>8)

		f.DrawString(string(v), pt)
		x = x + int(s.FontSize)
	}
	s.img = img
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

// VerifySecCode 不区分大小写
func VerifySecCode(seccode1, seccode2 string) bool {
	if len(seccode1) != len(seccode2) {
		return false
	}
	if !strings.EqualFold(seccode1, seccode2) {
		return false
	}
	return true
}

// VerifySecCodeIgnoreCase 区分大小写
func VerifySecCodeIgnoreCase(seccode1, seccode2 string) bool {
	if len(seccode1) != len(seccode2) {
		return false
	}
	if seccode1 == seccode2 {
		return true
	}
	return false
}
