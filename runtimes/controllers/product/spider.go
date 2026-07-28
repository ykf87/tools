package product

import (
	"errors"
	"fmt"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/db/spider"
	"tools/runtimes/funcs"
	"tools/runtimes/requests"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var productHeader = [][]string{
	[]string{"spu需保证唯一,一行一个商品"},
	[]string{
		"spu",
		"商品图",
		"商品视频",
		"商家编码",
		"原价",
		"售价",
		"进货价",
		"分类",
		"标签",
		"sku",
		"sku图片",
		"sku视频",
		"sku价格",
		"属性",
		"筛选",
		"库存",
		"虚拟销量",
		"商品重量",
		"包裹重量",
		"长",
		"宽",
		"高",
		"上架",
		"上架时间",
	},
}

var i18nHeader = [][]string{
	[]string{"产品多语言信息,语言项需从后台语言管理处获取,spu需和Sheet1项一致"},
	[]string{
		"spu",
		"语言",
		"标题",
		"副标题",
		"简介",
		"详情",
		"seo标题",
		"seo简介",
		"关键词",
	},
}

var cateHeader = [][]string{
	[]string{"编码需保证在分类项中唯一"},
	[]string{
		"编码",
		"上级分类",
		"语言",
		"分类名",
		"分类简介",
		"seo标题",
		"seo简介",
		"关键词",
	},
}

var tagHeader = [][]string{
	[]string{"标签，编码需要保证在标签项中的唯一性"},
	[]string{
		"编码",
		"语言",
		"标签名称",
		"标签简介",
		"seo标题",
		"seo简介",
		"关键词",
	},
}
var attrHeader = [][]string{
	[]string{"单页如果设置为1,则必须填写单页标题和单页介绍,编码需保证在属性中唯一"},
	[]string{
		"编码",
		"语言",
		"属性名",
		"单页",
		"单页标题",
		"单页介绍",
		"单页关键词",
	},
}
var attrValueHeader = [][]string{
	[]string{"单页设置为1时,必须设置单页标题,单页简介和单页内容,用于将属性单独页面展示"},
	[]string{
		"编码",
		"属性",
		"下级属性",
		"下级筛选",
		"地区",
		"语言",
		"名称",
		"单页",
		"单页简介",
		"单页内容",
		"seo标题",
		"seo简介",
		"单页关键词",
		"图集",
		"视频",
		"虚拟喜欢",
	},
}

type Attr struct {
	Name   string   `json:"name" form:"name"`
	Images []string `json:"images" form:"images"`
	Videos []string `json:"videos" form:"videos"`
}

type ListData struct {
	Url      string   `json:"url" form:"url"`
	Name     string   `json:"name" form:"name"`
	SubName  string   `json:"sub_name" form:"sub_name"`
	Desc     string   `json:"desc" form:"desc"`
	SeoTitle string   `json:"seo_title" form:"seo_title"`
	SeoDesc  string   `json:"seo_desc" form:"seo_desc"`
	Spu      string   `jon:"spu" form:"spu"`
	Price    float64  `json:"price" form:"price"`
	Imgs     []string `json:"imgs" form:"imgs"`
	Videos   []string `json:"videos" form:"videos"`
	Attrs    []*Attr  `json:"attrs" form:"attrs"`
}

type ReqData struct {
	Domain   string      `json:"domain" form:"domain"`
	Cate     string      `json:"cate" form:"cate"`
	Tag      string      `json:"tag" form:"tag"`
	Brand    string      `json:"brand" form:"brand"`
	Lang     string      `json:"lang" form:"lang"`
	Attrname string      `json:"attrname" form:"attrname"`
	Lists    []*ListData `json:"lists" form:"lists"`
}

func GetSpiderJses(c *gin.Context) {
	response.Success(c, spider.GetSpiderJses(), "")
}

func SaveSpider(c *gin.Context) {
	var req spider.SpiderJs
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, 500, err.Error(), nil)
		return
	}

	if req.ID < 1 {
		req.CreatedAt = time.Now()
	} else {
		req.CreatedAt = time.Now()
	}

	if err := req.Save(&req, db.DB.DB()); err != nil {
		response.Error(c, 500, err.Error(), nil)
		return
	}

	response.Success(c, spider.GetSpiderJses(), "")
}

func Spider(c *gin.Context) {
	var req ReqData
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, 500, err.Error(), nil)
		return
	}

	if req.Lang == "" || req.Cate == "" {
		response.Error(c, 500, "请将右侧设置填写完整", nil)
		return
	}

	// 创建临时保存目录
	rootDir := filepath.Join("./data/tmp", funcs.RoundmUuid())
	if _, err := os.Stat(rootDir); err != nil {
		if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
			response.Error(c, 500, err.Error(), nil)
			return
		}
	}
	defer func() {
		// os.RemoveAll(rootDir)
		fmt.Println("要删除临时文件")
	}()

	var xml Nnp

	cateSheet := fmtCate(req.Cate, req.Lang)
	xml = append(xml, &XMLSData{
		SheetName: "分类",
		Datas:     cateSheet,
	})
	if req.Tag != "" {
		tagSheet := fmtTag(req.Tag, req.Lang)
		if len(tagSheet) > 0 {
			xml = append(xml, &XMLSData{
				SheetName: "标签",
				Datas:     tagSheet,
			})
		}
	}

	ps, pls, as, avs, err := fmtPros(rootDir, req)
	if err != nil {
		response.Error(c, 500, err.Error(), nil)
		return
	}
	xml = append(xml, &XMLSData{
		SheetName: "产品",
		Datas:     ps,
	}, &XMLSData{
		SheetName: "i18n",
		Datas:     pls,
	})
	if len(as) > 0 {
		xml = append(xml, &XMLSData{
			SheetName: "属性",
			Datas:     as,
		})
	}
	if len(avs) > 0 {
		xml = append(xml, &XMLSData{
			SheetName: "属性值",
			Datas:     avs,
		})
	}

	if err := xml.Output(rootDir); err != nil {
		response.Error(c, 500, err.Error(), nil)
		return
	}
	response.Error(c, 0, "success", nil)
}

// 下载文件
func downloadMedia(src, savePath, filename string) (string, error) {

	cli, err := requests.New(nil)
	if err != nil {
		return "", err
	}

	body, hd, err := cli.Get(src, nil)
	if err != nil {
		return "", err
	}

	if filename == "" {
		tttt := strings.Split(src, "?")
		src = tttt[0]
		imgSp := strings.Split(src, "/")
		filename = fmt.Sprintf("%s", imgSp[len(imgSp)-1])

		if !strings.Contains(filename, ".") {
			var contentType string
			for k, v := range hd {
				switch strings.ToLower(k) {
				case "content-type":
					if len(v) > 0 {
						contentType = v[0]
						goto ENDGETCT
					}
				}
			}
		ENDGETCT:

			if contentType == "" {
				fmt.Println("在返回头中无法确定文件格式")
				contentType = http.DetectContentType(body)
			} else {
				fmt.Println("返回头格式:", contentType)
			}

			exts, err := mime.ExtensionsByType(contentType)
			if err != nil {
				return "", err
			}
			if len(exts) < 1 {
				return "", errors.New("文件后缀无法确定!")
			}
			filename = filename + exts[0]
		}
	}

	if _, err := os.Stat(savePath); err != nil {
		if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
			return "", err
		}
	}

	return filename, os.WriteFile(filepath.Join(savePath, filename), body, 0644)

}

// 生成产品
func fmtPros(rootDir string, req ReqData) ([]*OutputProduct, []*OutputProductI18n, []*OutputAttribute, []*OutputAttributeValue, error) {
	var ProsSheet []*OutputProduct
	var ProI18nSheet []*OutputProductI18n
	var AttrSheet []*OutputAttribute
	var AttrValueSheet []*OutputAttributeValue
	imgDir := "images"
	imgRoot := filepath.Join(rootDir, imgDir)
	if _, err := os.Stat(imgRoot); err != nil {
		if err := os.MkdirAll(imgRoot, os.ModePerm); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	attrValueMap := make(map[string]bool)

	for idx, pro := range req.Lists {
		// if idx >= 1 {
		// 	break
		// }
		if pro.Spu == "" || pro.Price <= 0 || pro.Name == "" || len(pro.Imgs) < 1 {
			return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品数据不正确!", idx)
		}
		setAttrd := false
		var skus []string

		skuimg := make(map[string][]string)
		skuvideo := make(map[string][]string)
		if len(pro.Attrs) > 0 {
			if req.Attrname == "" {
				return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品 存在属性,但是未设置属性名称", idx)
			}
			req.Attrname = strings.Trim(req.Attrname, " ")

			if setAttrd == false {
				AttrSheet = append(AttrSheet, &OutputAttribute{
					Code:  req.Attrname,
					Name:  req.Attrname,
					Lang:  req.Lang,
					Title: req.Attrname,
				})
				setAttrd = true
			}

			for _, v := range pro.Attrs {
				v.Name = strings.Trim(v.Name, " ")
				if _, ok := attrValueMap[v.Name]; ok {
					continue
				}
				attrValueMap[v.Name] = true
				var imgs []string
				var videos []string
				if len(v.Images) > 0 {
					for _, img := range v.Images {
						if !strings.HasPrefix(strings.ToLower(img), "http") {
							if req.Domain == "" {
								return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品属性图片地址不含完整域名,请设置域名!", idx)
							}
							img = fmt.Sprintf("%s/%s", strings.TrimRight(req.Domain, "/"), strings.TrimLeft(img, "/"))
						}

						fn, err := downloadMedia(img, filepath.Join(imgRoot, pro.Spu), "")
						if err != nil {
							return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品图片下载错误: %s", idx, err.Error())
						}
						imgs = append(imgs, filepath.Join(imgDir, pro.Spu, fn))
					}
					skuimg[v.Name] = imgs
				}

				if len(v.Videos) > 0 {
					for _, video := range v.Videos {
						if !strings.HasPrefix(strings.ToLower(video), "http") {
							if req.Domain == "" {
								return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品属性图片地址不含完整域名,请设置域名!", idx)
							}
							video = fmt.Sprintf("%s/%s", strings.TrimRight(req.Domain, "/"), strings.TrimLeft(video, "/"))
						}

						fn, err := downloadMedia(video, filepath.Join(imgRoot, pro.Spu), "")
						if err != nil {
							return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品视频下载错误: %s", idx, err.Error())
						}
						videos = append(videos, filepath.Join(imgDir, pro.Spu, fn))
					}
					skuvideo[v.Name] = videos
				}
				skus = append(skus, v.Name)

				AttrValueSheet = append(AttrValueSheet, &OutputAttributeValue{
					Code:     v.Name,
					AttrName: req.Attrname,
					Lang:     req.Lang,
					Name:     v.Name,
					Title:    v.Name,
					Images:   imgs,
					Videos:   videos,
				})
			}
		}

		var proImgs []string
		var proVids []string
		for _, img := range pro.Imgs {
			if !strings.HasPrefix(strings.ToLower(img), "http") {
				if req.Domain == "" {
					return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品图片地址不含完整域名,请设置域名!", idx)
				}
				img = fmt.Sprintf("%s/%s", strings.TrimRight(req.Domain, "/"), strings.TrimLeft(img, "/"))
			}
			fn, err := downloadMedia(img, filepath.Join(imgRoot, pro.Spu), "")
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品图片下载错误: %s", idx, err.Error())
			}

			proImgs = append(proImgs, filepath.Join(imgDir, pro.Spu, fn))
		}
		if len(pro.Videos) > 0 {
			for _, video := range pro.Videos {
				if !strings.HasPrefix(strings.ToLower(video), "http") {
					if req.Domain == "" {
						return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品视频地址不含完整域名,请设置域名!", idx)
					}
					video = fmt.Sprintf("%s/%s", strings.TrimRight(req.Domain, "/"), strings.TrimLeft(video, "/"))
				}

				fn, err := downloadMedia(video, filepath.Join(imgRoot, pro.Spu), "")
				if err != nil {
					return nil, nil, nil, nil, fmt.Errorf("第 %d 个产品图片下载错误: %s", idx, err.Error())
				}

				proVids = append(proVids, filepath.Join(imgDir, pro.Spu, fn))
			}
		}

		skustr := ""
		if len(skus) > 0 {
			skustr = fmt.Sprintf("%s:%s", req.Attrname, strings.Join(skus, ","))
		}

		var skuimgStr string
		if len(skuimg) > 0 {
			var vvvvv []string
			for n, mmgs := range skuimg {
				vvvvv = append(vvvvv, fmt.Sprintf("%s:%s", n, strings.Join(mmgs, ",")))
			}
			skuimgStr = strings.Join(vvvvv, "\n")
		}

		var skuvidStr string
		if len(skuvideo) > 0 {
			var vvvvv []string
			for n, mmgs := range skuvideo {
				vvvvv = append(vvvvv, fmt.Sprintf("%s:%s", n, strings.Join(mmgs, ",")))
			}
			skuvidStr = strings.Join(vvvvv, "\n")
		}
		ProsSheet = append(ProsSheet, &OutputProduct{
			Spu:         pro.Spu,
			Images:      proImgs,
			Videos:      proVids,
			OriginPrice: pro.Price,
			SalePrice:   pro.Price,
			Cates:       req.Cate,
			Tags:        req.Tag,
			Sku:         skustr,
			SkuImages:   skuimgStr,
			SkuVideos:   skuvidStr,
		})
		ProI18nSheet = append(ProI18nSheet, &OutputProductI18n{
			Spu:      pro.Spu,
			Lang:     req.Lang,
			Title:    pro.Name,
			SubTitle: fmt.Sprintf("%s %s", pro.SeoTitle, pro.Spu),
			SeoTitle: pro.SeoTitle,
			Desc:     pro.Desc,
			SeoDesc:  pro.SeoDesc,
		})
	}
	return ProsSheet, ProI18nSheet, AttrSheet, AttrValueSheet, nil
}

// 生成标签表单
func fmtTag(tag, langStr string) []*OutputTag {
	tag = strings.ReplaceAll(tag, "，", ",")
	tags := strings.Split(tag, ",")

	var tagArr []*OutputTag
	for _, t := range tags {
		t = strings.Trim(t, " ")
		if t != "" {
			tagArr = append(tagArr, &OutputTag{
				Code:     t,
				Lang:     langStr,
				Name:     t,
				SeoTitle: t,
			})
		}
	}
	return tagArr
}

// 生成分类的表单
func fmtCate(cate, langstr string) []*OutputCate {
	cate = strings.ReplaceAll(cate, "，", ",")
	cate = strings.ReplaceAll(cate, "》", ">")
	cates := strings.Split(cate, ",")

	var cateArr []*OutputCate
	for _, c := range cates {
		cs := strings.Split(c, ">")
		var parent string
		for _, v := range cs {
			cateArr = append(cateArr, &OutputCate{
				Code:     v,
				Parent:   parent,
				Lang:     langstr,
				Name:     v,
				SeoTitle: v,
			})
		}
	}
	return cateArr
}

// 导出
type XMLSData struct {
	SheetName string `json:"sheet_name"`
	Datas     any    `json:"datas"`
}
type OutputProductI18n struct {
	Spu      string `json:"spu"`
	Lang     string `json:"lang"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	Desc     string `json:"desc"`
	Content  string `json:"content"`
	SeoTitle string `json:"seo_title"`
	SeoDesc  string `json:"seo_desc"`
	Keyword  string `json:"keyword"`
}
type OutputProduct struct {
	Spu         string   `json:"spu"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
	OriginPrice float64  `json:"origin_price"`
	SalePrice   float64  `json:"sale_price"`
	Cates       string   `json:"cates"`
	Tags        string   `json:"tags"`
	Sku         string   `json:"sku"`
	SkuImages   string   `json:"sku_images"`
	SkuVideos   string   `json:"sku_videos"`
	// Lang        map[string]OutputProductI18n `json:"lang"`
}

type OutputCate struct {
	Code     string `json:"code"`
	Parent   string `json:"parent"`
	Lang     string `json:"lang"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	SeoTitle string `json:"seo_title"`
	SeoDesc  string `json:"seo_desc"`
	Keyword  string `json:"keyword"`
}

type OutputTag struct {
	Code     string `json:"code"`
	Lang     string `json:"lang"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	SeoTitle string `json:"seo_title"`
	SeoDesc  string `json:"seo_desc"`
	Keyword  string `json:"keyword"`
}

type OutputAttribute struct {
	Code    string `json:"code"`
	Lang    string `json:"lang"`
	Name    string `json:"name"`
	IsPage  bool   `json:"is_page"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Keyword string `json:"keyword"`
}

type OutputAttributeValue struct {
	Code             string   `json:"code"`
	AttrName         string   `json:"attr_name"`
	ChildValueName   string   `json:"child_value_name"`
	ChildValueFilter string   `json:"child_value_filter"`
	Region           string   `json:"region"`
	Lang             string   `json:"lang"`
	Name             string   `json:"name"`
	IsPage           bool     `json:"is_page"`
	Title            string   `json:"title"`
	Desc             string   `json:"desc"`
	SeoTitle         string   `json:"seo_title"`
	SeoDesc          string   `json:"seo_desc"`
	Keyword          string   `json:"keyword"`
	Images           []string `json:"images"`
	Videos           []string `json:"videos"`
	VirtualLike      int      `json:"virtual_like"`
}

type Nnp []*XMLSData

func (xx Nnp) Output(rootDir string) error {
	f := excelize.NewFile()
	for _, v := range xx {
		if dt, ok := v.Datas.([]*OutputProduct); ok {
			if err := v.genProduct(f, dt); err != nil {
				return err
			}
		}
		if dt, ok := v.Datas.([]*OutputProductI18n); ok {
			if err := v.genProductI18n(f, dt); err != nil {
				return err
			}
		}
		if dt, ok := v.Datas.([]*OutputCate); ok {
			if err := v.genCates(f, dt); err != nil {
				return err
			}
		}
		if dt, ok := v.Datas.([]*OutputTag); ok {
			if err := v.genTag(f, dt); err != nil {
				return err
			}
		}
		if dt, ok := v.Datas.([]*OutputAttribute); ok {
			if err := v.genAttr(f, dt); err != nil {
				return err
			}
		}
		if dt, ok := v.Datas.([]*OutputAttributeValue); ok {
			if err := v.genAttrValue(f, dt); err != nil {
				return err
			}
		}
	}

	if err := f.SaveAs(filepath.Join(rootDir, "import.xlsx")); err != nil {
		return err
	}
	return nil
}

// 通用设置
// type XMLHeaderCol struct{

// }
//
//	type XMLHeaderData struct{
//		Values []XMLHeaderCol
//		Stype *excelize.Style
//	}
func (x *XMLSData) setHeader(f *excelize.File, dCols [][]string) error {
	maxIdx := 0
	rowlen := len(dCols)
	for _, v := range dCols {
		cl := len(v)
		if maxIdx < cl {
			maxIdx = cl
		}
	}

	maxColName, err := excelize.ColumnNumberToName(maxIdx)
	if err != nil {
		return err
	}

	for spanID, v := range dCols {
		cl := len(v)
		row := spanID + 1
		if cl < maxIdx {
			err := f.MergeCell(x.SheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%s%d", maxColName, row))
			if err != nil {
				return err
			}
		}

		for colIdx, colVal := range v {
			name, _ := excelize.ColumnNumberToName(colIdx + 1)
			if err := f.SetCellValue(x.SheetName, fmt.Sprintf("%s%d", name, row), colVal); err != nil {
				return err
			}
		}
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  16,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#1F4E78"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(x.SheetName, "A1", fmt.Sprintf("%s1", maxColName), titleStyle)
	f.SetRowHeight(x.SheetName, 1, 50)
	f.SetRowHeight(x.SheetName, 2, 40)

	f.SetPanes(x.SheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		YSplit:      rowlen,
		TopLeftCell: fmt.Sprintf("A%d", rowlen+1),
		ActivePane:  "bottomLeft",
	})

	return nil
}

func (x *XMLSData) setRow(f *excelize.File, rowIdx int, val []any) error {
	for idx, v := range val {
		name, _ := excelize.ColumnNumberToName(idx + 1)
		if err := f.SetCellValue(x.SheetName, fmt.Sprintf("%s%d", name, rowIdx), v); err != nil {
			return err
		}
	}
	return nil
}

func (x *XMLSData) genProduct(f *excelize.File, dts []*OutputProduct) error {
	if _, err := f.NewSheet(x.SheetName); err != nil {
		return err
	}

	if err := x.setHeader(f, productHeader); err != nil {
		return err
	}

	startCol := len(productHeader) + 1
	for idx, v := range dts {
		cates := strings.Split(v.Cates, ",")
		var cts []string
		for _, ccc := range cates {
			cns := strings.Split(strings.Trim(ccc, " "), ">")
			if len(cns) > 0 {
				cts = append(cts, cns[len(cns)-1])
			}
		}

		rowData := []any{
			v.Spu,
			strings.ReplaceAll(strings.Join(v.Images, ","), "\\", "/"),
			strings.ReplaceAll(strings.Join(v.Videos, ","), "\\", "/"),
			"",
			v.OriginPrice,
			v.SalePrice,
			"",
			strings.Join(cts, ","),
			v.Tags,
			v.Sku,
			v.SkuImages,     // sku图片
			v.SkuVideos,     // sku视频
			"",              // sku价格
			"",              // 属性
			"",              // 筛选
			"",              // 库存
			rand.Intn(9999), // 虚拟销量
			"",              // 商品重量
			"",              // 包裹重量
			"",              // 长
			"",              // 宽
			"",              // 高
			"1",             // 上架
			"",              // 上架时间
		}
		if err := x.setRow(f, startCol+idx, rowData); err != nil {
			return err
		}
	}

	return nil
}

func (x *XMLSData) genProductI18n(f *excelize.File, dts []*OutputProductI18n) error {
	if _, err := f.NewSheet(x.SheetName); err != nil {
		return err
	}

	if err := x.setHeader(f, i18nHeader); err != nil {
		return err
	}

	startCol := len(i18nHeader) + 1
	for idx, v := range dts {
		rowData := []any{
			v.Spu,
			v.Lang,
			v.Title,
			v.SubTitle,
			v.Desc,
			v.Content,
			v.SeoTitle,
			v.SeoDesc,
			v.Keyword,
		}
		if err := x.setRow(f, startCol+idx, rowData); err != nil {
			return err
		}
	}
	return nil
}

func (x *XMLSData) genCates(f *excelize.File, dts []*OutputCate) error {
	if _, err := f.NewSheet(x.SheetName); err != nil {
		return err
	}

	if err := x.setHeader(f, cateHeader); err != nil {
		return err
	}

	startCol := len(cateHeader) + 1
	for idx, v := range dts {
		rowData := []any{
			v.Code,
			v.Parent,
			v.Lang,
			v.Name,
			v.Desc,
			v.SeoTitle,
			v.SeoDesc,
			v.Keyword,
		}
		if err := x.setRow(f, startCol+idx, rowData); err != nil {
			return err
		}
	}
	return nil
}

func (x *XMLSData) genTag(f *excelize.File, dts []*OutputTag) error {
	if _, err := f.NewSheet(x.SheetName); err != nil {
		return err
	}

	if err := x.setHeader(f, tagHeader); err != nil {
		return err
	}

	startCol := len(tagHeader) + 1
	for idx, v := range dts {
		rowData := []any{
			v.Code,
			v.Lang,
			v.Name,
			v.Desc,
			v.SeoTitle,
			v.SeoDesc,
			v.Keyword,
		}
		if err := x.setRow(f, startCol+idx, rowData); err != nil {
			return err
		}
	}
	return nil
}

func (x *XMLSData) genAttr(f *excelize.File, dts []*OutputAttribute) error {
	if _, err := f.NewSheet(x.SheetName); err != nil {
		return err
	}

	if err := x.setHeader(f, attrHeader); err != nil {
		return err
	}

	startCol := len(attrHeader) + 1
	for idx, v := range dts {
		rowData := []any{
			v.Code,
			v.Lang,
			v.Name,
			"",
			v.Title,
			v.Desc,
			v.Keyword,
		}
		if err := x.setRow(f, startCol+idx, rowData); err != nil {
			return err
		}
	}
	return nil
}

func (x *XMLSData) genAttrValue(f *excelize.File, dts []*OutputAttributeValue) error {
	if _, err := f.NewSheet(x.SheetName); err != nil {
		return err
	}

	if err := x.setHeader(f, attrValueHeader); err != nil {
		return err
	}

	startCol := len(attrValueHeader) + 1
	for idx, v := range dts {
		rowData := []any{
			v.Code,
			v.AttrName,
			v.ChildValueName,
			v.ChildValueFilter,
			v.Region,
			v.Lang,
			v.Name,
			"", //单页
			v.Desc,
			"",
			v.SeoTitle,
			v.SeoDesc,
			v.Keyword,
			strings.Join(v.Images, ","),
			strings.Join(v.Videos, ","),
			v.VirtualLike,
		}
		if err := x.setRow(f, startCol+idx, rowData); err != nil {
			return err
		}
	}
	return nil
}
