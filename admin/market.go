package admin

import (
	"chat/globals"
	"chat/utils"
	"fmt"

	"github.com/spf13/viper"
)

type ModelTag []string

type MarketModel struct {
	Id              string   `json:"id" mapstructure:"id" yaml:"id" required:"true"`
	Name            string   `json:"name" mapstructure:"name" yaml:"name" required:"true"`
	Description     string   `json:"description" mapstructure:"description" yaml:"description"`
	Default         bool     `json:"default" mapstructure:"default" yaml:"default"`
	HighContext     bool     `json:"high_context" mapstructure:"highcontext" yaml:"highcontext"`
	FunctionCalling bool     `json:"function_calling" mapstructure:"functioncalling" yaml:"functioncalling"`
	VisionModel     bool     `json:"vision_model" mapstructure:"visionmodel" yaml:"visionmodel"`
	ThinkingModel   bool     `json:"thinking_model" mapstructure:"thinkingmodel" yaml:"thinkingmodel"`
	ReverseModel    bool     `json:"reverse_model" mapstructure:"reversemodel" yaml:"reversemodel"`
	OCRModel        bool     `json:"ocr_model" mapstructure:"ocrmodel" yaml:"ocrmodel"`
	Avatar          string   `json:"avatar" mapstructure:"avatar" yaml:"avatar"`
	Tag             ModelTag `json:"tag" mapstructure:"tag" yaml:"tag"`
}

type MarketModelList []MarketModel

type Market struct {
	Models MarketModelList `json:"models" mapstructure:"models"`
}

func NewMarket() *Market {
	var models MarketModelList

	if err := viper.UnmarshalKey("market", &models); err != nil {
		globals.Warn(
			fmt.Sprintf(
				"[market] read config error: %s, use default config",
				err.Error(),
			),
		)
		models = MarketModelList{}
	}

	return &Market{
		Models: models,
	}
}

func (m *Market) GetModels() MarketModelList {
	return m.Models
}

func (m *Market) GetModel(id string) *MarketModel {
	for _, model := range m.Models {
		if model.Id == id {
			return &model
		}
	}

	return nil
}

func (m *Market) SaveConfig() error {
	return utils.SaveConfig("market", m.Models)
}

func (m *Market) SetModels(models MarketModelList) error {
	m.Models = models

	return m.SaveConfig()
}
