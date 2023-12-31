package libraries

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/Raihanpoke/fullstack/config"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	en_translation "github.com/go-playground/validator/v10/translations/en"
)

type Validation struct {
	db *sql.DB
}

func NewValidation() *Validation {
	db, err := config.DBConnection()
	if err != nil {
		panic(err)
	}

	return &Validation{
		db: db,
	}
}

func (v *Validation) Init() (*validator.Validate, ut.Translator) {
	// memanggil package translator
	translator := en.New()
	uni := ut.New(translator, translator)

	trans, _ := uni.GetTranslator("en")

	validate := validator.New()

	// register default translation (en)
	en_translation.RegisterDefaultTranslations(validate, trans)

	// mengubah label default
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		labelName := field.Tag.Get("label")
		return labelName
	})

	// menambah default bahasa indonesia pada tag required
	validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} tidak boleh kosong", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		fmt.Println(t)
		return t
	})

	validate.RegisterValidation("isunique", func(fl validator.FieldLevel) bool {
		params := fl.Param()
		split_params := strings.Split(params, "-")

		tableName := split_params[0]
		fieldName := split_params[1]
		fieldValue := fl.Field().String()

		return v.CheckIsUnique(tableName, fieldName, fieldValue)
	})

	validate.RegisterTranslation("isunique", trans, func(ut ut.Translator) error {
		return ut.Add("isunique", "{0} sudah digunakan", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("isunique", fe.Field())
		return t
	})

	return validate, trans
}

func (v *Validation) Struct(s interface{}) interface{} {

	validate, trans := v.Init()

	vErrors := make(map[string]interface{})

	err := validate.Struct(s)

	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			vErrors[e.StructField()] = e.Translate(trans)
		}
	}

	if len(vErrors) > 0 {
		return vErrors
	}

	return nil
}

func (v *Validation) CheckIsUnique(tableName, fieldName, fieldValue string) bool {

	row, err := v.db.Query("select "+fieldName+" from "+tableName+" where "+fieldName+" = ?", fieldValue)
	if err != nil {
		panic(err)
	}
	defer row.Close()

	var result string
	for row.Next() {
		row.Scan(&result)
	}

	// email@tentangkode
	// email@tentangkode

	return result != fieldValue

}
