package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"embed"
	"path/filepath"
	"log"

	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tomnomnom/xtermcolor"
	"golang.org/x/term"
)

//go:embed static
var staticFiles embed.FS

//--- -s-
type Encoder struct {
	w           io.Writer
	isColor     bool
	Width       int
	Height      int
}
//
func NewEncoder(w io.Writer, isColor bool) *Encoder {
	return &Encoder{
		w: w, 
		isColor: isColor,
	}
}
//
func (e *Encoder) Encode(img image.Image) error {
	/* structのwにSixel形式にデコード後のデータを書き込み */
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w == 0 || h == 0 {
		return nil
	}
	terminalW, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && terminalW > 0 && w > terminalW*ansimage.BlockSizeX {
		w = terminalW * ansimage.BlockSizeX
	}
	if e.Width != 0 {
		w = e.Width
	}
	if e.Height != 0 {
		h = e.Height
	}
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		return err
	}

	sm := ansimage.ScaleMode(2)
	dm := ansimage.DitheringMode(0)
	mc, err := colorful.Hex("#000000")
	if err != nil {
		return err
	}
	pix, err := ansimage.NewScaledFromReader(&imgBuf, h/4, w/4, mc, sm, dm)
	if err != nil {
		return err
	}
	decodedSixel := pix.Render()
	if e.isColor {
		decoded8BitColor := to8BitColor(decodedSixel)
		e.w.Write([]byte(decoded8BitColor)) //コマンドメソッド structのwに書き込み
	} else {
		e.w.Write([]byte(decodedSixel)) //コマンドメソッド structのwに書き込み
	}
	return nil
}
// private関数
func to8BitColor(sixel string) string {
	re := regexp.MustCompile(`\x1b\[([34]8);2;(\d+);(\d+);(\d+)((;\d+)*)m`)
	found := re.FindAllStringSubmatchIndex(sixel, -1)
	pos := 0
	var builder strings.Builder
	for i := 0; i < len(found); i++ {
		if pos < found[i][0] {
			builder.WriteString(sixel[pos:found[i][0]])
		}
		r, _ := strconv.Atoi(sixel[found[i][4]:found[i][5]])
		g, _ := strconv.Atoi(sixel[found[i][6]:found[i][7]])
		b, _ := strconv.Atoi(sixel[found[i][8]:found[i][9]])
		c := xtermcolor.FromColor(color.RGBA{uint8(r), uint8(g), uint8(b), 255})
		builder.WriteString(fmt.Sprintf("\x1b[%s;5;%d%sm", sixel[found[i][2]:found[i][3]], c, sixel[found[i][10]:found[i][11]]))
		pos = found[i][1]
	}
	builder.WriteString(sixel[pos:])
	return builder.String()
}
//---e-

func main() {
	// githubで使うにはembedでフォルダの指定必須
	imgPath := filepath.ToSlash(filepath.Join("static", "animal.png"))

	file, err := staticFiles.Open(imgPath) //(os.Open()はNG)迷子対策
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	originalImg, err := png.Decode(file) //pngで動作(os.Decode()はNG)
	if err != nil {
		log.Fatal(err)
	}

	//標準出力（stdout）
	isColor := true
	stdoutEncoder := NewEncoder(os.Stdout, isColor)
	err = stdoutEncoder.Encode(originalImg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n里親募集 スコティッシュフォールドクリームタビー血統種 5歳オス去勢済みの家猫です。")
	fmt.Println("引っ掻き癖や噛み癖無しのおとなしい性格です。")
	fmt.Println("住居は川崎駅周辺です。2024年8月引っ越し予定")
	fmt.Println("里親募集理由: 親の介護のため実家に戻るのですが親が猫NGのため。")

}