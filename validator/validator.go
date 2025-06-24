package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	id_translations "github.com/go-playground/validator/v10/translations/id"
)

var (
	Validate     *validator.Validate
	TranslatorID ut.Translator
	TranslatorEn ut.Translator
)

func InitValidator() {
	Validate = validator.New()
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "-" {
			return ""
		}

		name := strings.SplitN(tag, ",", 2)[0]
		if name == "" {
			return fld.Name
		}
		return name
	})

	// Setup indonesia translator
	indonesia := id.New()
	uni := ut.New(indonesia, indonesia)

	var found bool
	TranslatorID, found = uni.GetTranslator("id")
	if !found {
		panic("translator not found")
	}

	err := id_translations.RegisterDefaultTranslations(Validate, TranslatorID)
	if err != nil {
		panic(err)
	}

	english := en.New()
	uniEn := ut.New(english, english)

	TranslatorEn, found = uniEn.GetTranslator("en")
	if !found {
		panic("translator not found")
	}

	err = en_translations.RegisterDefaultTranslations(Validate, TranslatorEn)
	if err != nil {
		panic(err)
	}

}
