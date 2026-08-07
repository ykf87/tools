package product

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
		"品牌",
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

type OutputData struct {
	CateNames string `json:"cate_names"`
	Tags      string `json:"tags"`
	Lang      string `json:"lang"`
}

// ProsSheet = append(ProsSheet, &OutputProduct{
// 			Spu:         pro.Spu,
// 			Images:      proImgs,
// 			Videos:      proVids,
// 			OriginPrice: pro.Price,
// 			SalePrice:   fmtSalePrice(req.Prices, pro.Price),
// 			Cates:       req.Cate,
// 			Tags:        req.Tag,
// 			Sku:         skustr,
// 			SkuImages:   skuimgStr,
// 			SkuVideos:   skuvidStr,
// 			Brand:       req.Brand,
// 		})
// ProI18nSheet = append(ProI18nSheet, &OutputProductI18n{
// 			Spu:      pro.Spu,
// 			Lang:     req.Lang,
// 			Title:    pro.Name,
// 			SubTitle: fmt.Sprintf("%s %s", pro.SeoTitle, pro.Spu),
// 			SeoTitle: pro.SeoTitle,
// 			Desc:     pro.Desc,
// 			SeoDesc:  pro.SeoDesc,
// 		})
//
// AttrSheet = append(AttrSheet, &OutputAttribute{
// 	Code:  req.Attrname,
// 	Name:  req.Attrname,
// 	Lang:  req.Lang,
// 	Title: req.Attrname,
// })
//
//
// AttrValueSheet = append(AttrValueSheet, &OutputAttributeValue{
// 	Code:     v.Name,
// 	AttrName: req.Attrname,
// 	Lang:     req.Lang,
// 	Name:     v.Name,
// 	Title:    v.Name,
// 	Images:   imgs,
// 	Videos:   videos,
// })

func Output() error {
	// rootDir := filepath.Join("./data/tmp", funcs.RoundmUuid())
	// if _, err := os.Stat(rootDir); err != nil {
	// 	if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
	// 		return err
	// 	}
	// }
	// defer func() {
	// 	// os.RemoveAll(rootDir)
	// 	fmt.Println("要删除临时文件")
	// }()

	// var xml Nnp

	// cateSheet := fmtCate(req.Cate, req.Lang)
	// xml = append(xml, &XMLSData{
	// 	SheetName: "分类",
	// 	Datas:     cateSheet,
	// })
	// if req.Tag != "" {
	// 	tagSheet := fmtTag(req.Tag, req.Lang)
	// 	if len(tagSheet) > 0 {
	// 		xml = append(xml, &XMLSData{
	// 			SheetName: "标签",
	// 			Datas:     tagSheet,
	// 		})
	// 	}
	// }

	// ps, pls, as, avs, err := fmtPros(rootDir, req)
	// if err != nil {
	// 	fmt.Println(err, "格式化产品失败")

	// 	return err
	// }
	// xml = append(xml, &XMLSData{
	// 	SheetName: "产品",
	// 	Datas:     ps,
	// }, &XMLSData{
	// 	SheetName: "i18n",
	// 	Datas:     pls,
	// })
	// if len(as) > 0 {
	// 	xml = append(xml, &XMLSData{
	// 		SheetName: "属性",
	// 		Datas:     as,
	// 	})
	// }
	// if len(avs) > 0 {
	// 	xml = append(xml, &XMLSData{
	// 		SheetName: "属性值",
	// 		Datas:     avs,
	// 	})
	// }

	// if err := xml.Output(rootDir); err != nil {
	// 	fmt.Println(err, "导出xmsl失败")
	// 	return err
	// }
	return nil
}
