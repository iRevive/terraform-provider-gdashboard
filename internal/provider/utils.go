package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iRevive/terraform-provider-gdashboard/internal/provider/grafana"
	"hash/crc32"
)

// defaults

type FieldDefaults struct {
	Unit       string
	Decimals   *int
	Min        *float64
	Max        *float64
	NoValue    *float64
	Color      ColorDefaults
	Thresholds ThresholdDefaults
}

func NewFieldDefaults() FieldDefaults {
	return FieldDefaults{
		Unit:     "",
		Decimals: nil,
		Min:      nil,
		Max:      nil,
		Color: ColorDefaults{
			Mode:       "palette-classic",
			FixedColor: "green",
			SeriesBy:   "last",
		},
		Thresholds: ThresholdDefaults{
			Mode: "absolute",
			Steps: []ThresholdStepDefaults{
				{
					Color: "green",
					Value: nil,
				},
			},
		},
	}
}

type ColorDefaults struct {
	Mode       string
	FixedColor string
	SeriesBy   string
}

type ThresholdDefaults struct {
	Mode  string
	Steps []ThresholdStepDefaults
}

type ThresholdStepDefaults struct {
	Color string
	Value *float64
}

type ReduceOptionDefaults struct {
	Values      bool
	Fields      string
	Limit       *int
	Calculation string
}

func NewReduceOptionDefaults() ReduceOptionDefaults {
	return ReduceOptionDefaults{
		Values:      false,
		Fields:      "",
		Calculation: "lastNotNull",
	}
}

type TextSizeDefaults struct {
	Title *int
	Value *int
}

type AxisDefaults struct {
	Label     string
	Placement string
	SoftMin   *int
	SoftMax   *int
	Scale     ScaleDefaults
}

type ScaleDefaults struct {
	Type string
	Log  int
}

// Terraform projections

type AxisOptions struct {
	Label     types.String   `tfsdk:"label"`
	Placement types.String   `tfsdk:"placement"`
	SoftMin   types.Int64    `tfsdk:"soft_min"`
	SoftMax   types.Int64    `tfsdk:"soft_max"`
	Scale     []ScaleOptions `tfsdk:"scale"`
}

type ScaleOptions struct {
	Type types.String `tfsdk:"type"`
	Log  types.Int64  `tfsdk:"log"`
}

type MappingOptions struct {
	Value   []ValueMappingOptions   `tfsdk:"value"`
	Range   []RangeMappingOptions   `tfsdk:"range"`
	Regex   []RegexMappingOptions   `tfsdk:"regex"`
	Special []SpecialMappingOptions `tfsdk:"special"`
}

type ValueMappingOptions struct {
	Value       types.String `tfsdk:"value"`
	DisplayText types.String `tfsdk:"display_text"`
	Color       types.String `tfsdk:"color"`
}

type RangeMappingOptions struct {
	From        types.Number `tfsdk:"from"`
	To          types.Number `tfsdk:"to"`
	DisplayText types.String `tfsdk:"display_text"`
	Color       types.String `tfsdk:"color"`
}

type RegexMappingOptions struct {
	Pattern     types.String `tfsdk:"pattern"`
	DisplayText types.String `tfsdk:"display_text"`
	Color       types.String `tfsdk:"color"`
}

type SpecialMappingOptions struct {
	Match       types.String `tfsdk:"match"`
	DisplayText types.String `tfsdk:"display_text"`
	Color       types.String `tfsdk:"color"`
}

type FieldOptions struct {
	Unit       types.String       `tfsdk:"unit"`
	Decimals   types.Int64        `tfsdk:"decimals"`
	Min        types.Float64      `tfsdk:"min"`
	Max        types.Float64      `tfsdk:"max"`
	NoValue    types.Float64      `tfsdk:"no_value"`
	Color      []ColorOptions     `tfsdk:"color"`
	Mappings   []MappingOptions   `tfsdk:"mappings"`
	Thresholds []ThresholdOptions `tfsdk:"thresholds"`
	// todo links
}

type ColorOptions struct {
	Mode       types.String `tfsdk:"mode"`
	FixedColor types.String `tfsdk:"fixed_color"`
	SeriesBy   types.String `tfsdk:"series_by"`
}

type ThresholdOptions struct {
	Mode  types.String    `tfsdk:"mode"`
	Steps []ThresholdStep `tfsdk:"step"`
}

type ThresholdStep struct {
	Color types.String  `tfsdk:"color"`
	Value types.Float64 `tfsdk:"value"`
}

type ReduceOptions struct {
	Values      types.Bool   `tfsdk:"values"`
	Fields      types.String `tfsdk:"fields"`
	Limit       types.Int64  `tfsdk:"limit"`
	Calculation types.String `tfsdk:"calculation"`
}

type TextSizeOptions struct {
	Title types.Int64 `tfsdk:"title"`
	Value types.Int64 `tfsdk:"value"`
}

type Target struct {
	Prometheus []PrometheusTarget `tfsdk:"prometheus"`
	CloudWatch []CloudWatchTarget `tfsdk:"cloudwatch"`
}

type PrometheusTarget struct {
	Uid     types.String `tfsdk:"uid"`
	Expr    types.String `tfsdk:"expr"`
	Instant types.Bool   `tfsdk:"instant"`
	Format  types.String `tfsdk:"format"`
	// etc
	RefId        types.String `tfsdk:"ref_id"`
	MinInterval  types.String `tfsdk:"min_interval"`
	LegendFormat types.String `tfsdk:"legend_format"`
}

type CloudWatchTarget struct {
	Uid        types.String          `tfsdk:"uid"`
	Namespace  types.String          `tfsdk:"namespace"`
	MetricName types.String          `tfsdk:"metric_name"`
	Statistic  types.String          `tfsdk:"statistic"`
	Dimensions []CloudWatchDimension `tfsdk:"dimension"`
	MatchExact types.Bool            `tfsdk:"match_exact"`
	Region     types.String          `tfsdk:"region"`
	// etc
	RefId        types.String `tfsdk:"ref_id"`
	Period       types.String `tfsdk:"period"`
	LegendFormat types.String `tfsdk:"legend_format"`
}

type CloudWatchDimension struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func axisBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Description: "Axis display options.",
		Blocks: map[string]tfsdk.Block{
			"scale": {
				NestingMode: tfsdk.BlockNestingModeList,
				MinItems:    0,
				MaxItems:    1,
				Description: "Can be used to configure the scale of the y-axis.",
				MarkdownDescription: "Can be used to configure the scale of the y-axis. " +
					"Another way visualize series that differ by orders of magnitude is to use a logarithmic scales. " +
					"This is really useful for data usage or latency measurements. " +
					"The goal here is to avoid one series dominating and delegating all the others to the bottom of the graph.",
				Attributes: map[string]tfsdk.Attribute{
					"type": {
						Type:                types.StringType,
						Required:            true,
						Description:         "The type of the scale. The choices are: linear, log.",
						MarkdownDescription: "The type of the scale. The choices are: `linear`, `log`.",
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf("linear", "log"),
						},
					},
					"log": {
						Type:                types.Int64Type,
						Optional:            true,
						Description:         "The power of the logarithmic scale. The choices are: 2, 10.",
						MarkdownDescription: "The power of the logarithmic scale. The choices are: `2`, `10`.",
						Validators: []tfsdk.AttributeValidator{
							int64validator.OneOf(2, 10),
						},
					},
				},
			},
		},
		Attributes: map[string]tfsdk.Attribute{
			"label": {
				Type:        types.StringType,
				Description: "The custom text label for the y-axis.",
				Optional:    true,
			},
			"placement": {
				Type:                types.StringType,
				Optional:            true,
				Description:         "The placement of the y-axis. The choices are: auto, left, right, hidden.",
				MarkdownDescription: "The placement of the y-axis. The choices are: `auto`, `left`, `right`, `hidden`.",
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("auto", "left", "right", "hidden"),
				},
			},
			"soft_min": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "The soft minimum of y-axis.",
				MarkdownDescription: "The soft minimum of y-axis. " +
					"By default, the Grafana workspace sets the range for the y-axis automatically based on the data." +
					"The `soft_min` setting can prevent blips from appearing as mountains when the data is mostly flat, " +
					"and hard min or max derived from standard min and max field options can prevent intermittent spikes " +
					"from flattening useful detail by clipping the spikes past a defined point.",
			},
			"soft_max": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "The soft maximum of y-axis.",
				MarkdownDescription: "The soft maximum of y-axis. " +
					"By default, the Grafana workspace sets the range for the y-axis automatically based on the data." +
					"The `soft_max` setting can prevent blips from appearing as mountains when the data is mostly flat, " +
					"and hard min or max derived from standard min and max field options can prevent intermittent spikes " +
					"from flattening useful detail by clipping the spikes past a defined point.",
			},
			/*"width": {
				Type:     types.Int64Type,
				Optional: true,
				Description: "The fixed width of the y-axis.",
				MarkdownDescription: "The fixed width of the y-axis. By default, the Grafana workspace dynamically calculates the axis width. " +
					"By setting the width of the axis, data whose axes types are different can share the same display proportions. " +
					"This makes it easier to compare more than one graph’s worth of data because the axes are not shifted or stretched within visual proximity of each other.",
			},*/
		},
	}
}

func fieldBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Blocks: map[string]tfsdk.Block{
			"color": {
				NestingMode: tfsdk.BlockNestingModeList,
				MinItems:    0,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"mode": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf(
								"fixed", "thresholds", "palette-classic",
								"continuous-GrYlRd", "continuous-RdYlGr", "continuous-BlYlRd", "continuous-YlRd", "continuous-BlPu", "continuous-YlBl",
								"continuous-blues", "continuous-reds", "continuous-greens", "continuous-purples",
							),
						},
					},
					"fixed_color": {
						Type:     types.StringType,
						Optional: true,
					},
					"series_by": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
			"thresholds": {
				NestingMode: tfsdk.BlockNestingModeList,
				MinItems:    0,
				MaxItems:    1,
				Blocks: map[string]tfsdk.Block{
					"step": {
						NestingMode: tfsdk.BlockNestingModeList,
						MaxItems:    20,
						Attributes: map[string]tfsdk.Attribute{
							"color": {
								Type:     types.StringType,
								Required: true,
							},
							"value": {
								Type:     types.Float64Type,
								Optional: true,
							},
						},
					},
				},
				Attributes: map[string]tfsdk.Attribute{
					"mode": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf("absolute", "percentage"),
						},
					},
				},
			},
			"mappings": mappingsBlock(),
		},
		Attributes: map[string]tfsdk.Attribute{
			"unit": {
				Type:     types.StringType,
				Optional: true,
			},
			"decimals": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"min": {
				Type:     types.Float64Type,
				Optional: true,
			},
			"max": {
				Type:     types.Float64Type,
				Optional: true,
			},
			"no_value": {
				Type:     types.Float64Type,
				Optional: true,
			},
		},
	}
}

func reduceOptionsBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Attributes: map[string]tfsdk.Attribute{
			"values": {
				Type:     types.BoolType,
				Optional: true,
			},
			"fields": {
				Type:     types.StringType,
				Optional: true,
			},
			"limit": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"calculation": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}
}

func textSizeBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Attributes: map[string]tfsdk.Attribute{
			"title": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"value": {
				Type:     types.Int64Type,
				Optional: true,
			},
		},
	}
}

func targetBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MaxItems:    3,
		Blocks: map[string]tfsdk.Block{
			"prometheus": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    5,
				Attributes: map[string]tfsdk.Attribute{
					"uid": {
						Type:                types.StringType,
						MarkdownDescription: "Prometheus DataSource UID",
						Required:            true,
					},
					"expr": {
						Type:     types.StringType,
						Optional: false,
						Required: true,
					},
					"instant": {
						Type:     types.BoolType,
						Optional: true,
					},
					"ref_id": {
						Type:     types.StringType,
						Optional: true,
					},
					"format": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf("time_series", "table", "heatmap"),
						},
					},
					"min_interval": {
						Type:     types.StringType,
						Optional: true,
					},
					"legend_format": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
			"cloudwatch": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    5,
				Blocks: map[string]tfsdk.Block{
					"dimension": {
						NestingMode: tfsdk.BlockNestingModeList,
						MaxItems:    5,
						Attributes: map[string]tfsdk.Attribute{
							"name": {
								Type:     types.StringType,
								Required: true,
							},
							"value": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Attributes: map[string]tfsdk.Attribute{
					"uid": {
						Type:                types.StringType,
						MarkdownDescription: "CloudWatch DataSource UID",
						Required:            true,
					},
					"namespace": {
						Type:     types.StringType,
						Required: true,
					},
					"metric_name": {
						Type:     types.StringType,
						Required: true,
					},
					"statistic": {
						Type:     types.StringType,
						Required: true,
					},
					"match_exact": {
						Type:     types.BoolType,
						Optional: true,
					},
					"region": {
						Type:     types.StringType,
						Optional: true,
					},
					"ref_id": {
						Type:     types.StringType,
						Optional: true,
					},
					"period": {
						Type:     types.StringType,
						Optional: true,
					},
					"legend_format": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
		},
	}
}

func mappingsBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Blocks: map[string]tfsdk.Block{
			"value": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    10,
				Attributes: map[string]tfsdk.Attribute{
					"value": {
						Type:     types.StringType,
						Required: true,
					},
					"display_text": {
						Type:     types.StringType,
						Optional: true,
					},
					"color": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
			"range": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    10,
				Attributes: map[string]tfsdk.Attribute{
					"from": {
						Type:     types.StringType,
						Required: true,
					},
					"to": {
						Type:     types.StringType,
						Required: true,
					},
					"display_text": {
						Type:     types.StringType,
						Optional: true,
					},
					"color": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
			"regex": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    10,
				Attributes: map[string]tfsdk.Attribute{
					"pattern": {
						Type:     types.StringType,
						Required: true,
					},
					"display_text": {
						Type:     types.StringType,
						Optional: true,
					},
					"color": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
			"special": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    10,
				Attributes: map[string]tfsdk.Attribute{
					"match": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf("null", "nan", "null+nan", "true", "false", "empty"),
						},
					},
					"display_text": {
						Type:     types.StringType,
						Optional: true,
					},
					"color": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
		},
	}
}

// attributes
func idAttribute() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:     types.StringType,
		Computed: true,
	}
}

func jsonAttribute() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:        types.StringType,
		Computed:    true,
		Description: "The Grafana-API-compatible JSON of this panel.",
	}
}

func titleAttribute() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:        types.StringType,
		Required:    true,
		Description: "The title of this panel.",
	}
}

func descriptionAttribute() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:        types.StringType,
		Optional:    true,
		Description: "The description of this panel.",
	}
}

// creators

func createTargets(dataTargets []Target) []grafana.Target {
	targets := make([]grafana.Target, 0)

	for _, group := range dataTargets {
		for _, target := range group.Prometheus {
			t := grafana.Target{
				Datasource: grafana.Datasource{
					UID:  target.Uid.Value,
					Type: "prometheus",
				},
				RefID:        target.RefId.Value,
				Expr:         target.Expr.Value,
				Interval:     target.MinInterval.Value,
				LegendFormat: target.LegendFormat.Value,
				Instant:      target.Instant.Value,
				Format:       target.Format.Value,
			}

			targets = append(targets, t)
		}

		for _, target := range group.CloudWatch {
			dimensions := make(map[string]string)

			for _, dim := range target.Dimensions {
				dimensions[dim.Name.Value] = dim.Value.Value
			}

			t := grafana.Target{
				Datasource: grafana.Datasource{
					UID:  target.Uid.Value,
					Type: "cloudwatch",
				},
				RefID:      target.RefId.Value,
				Namespace:  target.Namespace.Value,
				MetricName: target.MetricName.Value,
				Statistics: []string{target.Statistic.Value},
				Dimensions: dimensions,
				Period:     target.Period.Value,
				Region:     target.Region.Value,
				Label:      target.LegendFormat.Value,
			}

			targets = append(targets, t)
		}
	}

	return targets
}

type ValueMappingResult struct {
	Color string `json:"color,omitempty"`
	Text  string `json:"text,omitempty"`
	Index int    `json:"index"`
}

func createFieldConfig(defaults FieldDefaults, fieldOptions []FieldOptions) grafana.FieldConfigDefaults {
	thresholdStep := make([]grafana.ThresholdStep, len(defaults.Thresholds.Steps))

	for i, step := range defaults.Thresholds.Steps {
		thresholdStep[i] = grafana.ThresholdStep{
			Color: step.Color,
			Value: step.Value,
		}
	}

	fieldConfig := grafana.FieldConfigDefaults{
		Unit:     defaults.Unit,
		Decimals: defaults.Decimals,
		Min:      defaults.Min,
		Max:      defaults.Max,
		Color: grafana.FieldConfigColor{
			Mode:       defaults.Color.Mode,
			FixedColor: defaults.Color.FixedColor,
			SeriesBy:   defaults.Color.SeriesBy,
		},
		Thresholds: grafana.Thresholds{
			Mode:  defaults.Thresholds.Mode,
			Steps: thresholdStep,
		},
	}

	for _, field := range fieldOptions {
		if !field.Unit.Null {
			fieldConfig.Unit = field.Unit.Value
		}

		if !field.Decimals.Null {
			decimals := int(field.Decimals.Value)
			fieldConfig.Decimals = &decimals
		}

		if !field.Min.Null {
			fieldConfig.Min = &field.Min.Value
		}

		if !field.Max.Null {
			fieldConfig.Max = &field.Max.Value
		}

		if !field.NoValue.Null {
			fieldConfig.NoValue = &field.NoValue.Value
		}

		for _, color := range field.Color {
			if !color.Mode.Null {
				fieldConfig.Color.Mode = color.Mode.Value
			}

			if !color.FixedColor.Null {
				fieldConfig.Color.FixedColor = color.FixedColor.Value
			}

			if !color.SeriesBy.Null {
				fieldConfig.Color.SeriesBy = color.SeriesBy.Value
			}
		}

		mappings := make([]grafana.FieldMapping, 0)

		for _, mapping := range field.Mappings {
			idx := 0
			valuesMap := make(map[string]interface{})

			for _, value := range mapping.Value {
				v := ValueMappingResult{
					Color: value.Color.Value,
					Text:  value.DisplayText.Value,
					Index: idx,
				}

				valuesMap[value.Value.Value] = v
				idx += 1
			}

			if len(valuesMap) > 0 {
				mapping := grafana.FieldMapping{
					Type:    "value",
					Options: valuesMap,
				}

				mappings = append(mappings, mapping)
			}

			for _, range_ := range mapping.Range {
				mapping := grafana.FieldMapping{
					Type: "range",
					Options: map[string]interface{}{
						"from": range_.From.Value,
						"to":   range_.From.Value,
						"result": ValueMappingResult{
							Color: range_.Color.Value,
							Text:  range_.DisplayText.Value,
							Index: idx,
						},
					},
				}
				idx += 1

				mappings = append(mappings, mapping)
			}

			for _, regex := range mapping.Regex {
				mapping := grafana.FieldMapping{
					Type: "regex",
					Options: map[string]interface{}{
						"pattern": regex.Pattern.Value,
						"result": ValueMappingResult{
							Color: regex.Color.Value,
							Text:  regex.DisplayText.Value,
							Index: idx,
						},
					},
				}
				idx += 1

				mappings = append(mappings, mapping)
			}

			for _, special := range mapping.Special {
				mapping := grafana.FieldMapping{
					Type: "special",
					Options: map[string]interface{}{
						"match": special.Match.Value,
						"result": ValueMappingResult{
							Color: special.Color.Value,
							Text:  special.DisplayText.Value,
							Index: idx,
						},
					},
				}
				idx += 1

				mappings = append(mappings, mapping)
			}
		}

		if len(mappings) > 0 {
			fieldConfig.Mappings = mappings
		}

		for _, threshold := range field.Thresholds {
			steps := make([]grafana.ThresholdStep, len(threshold.Steps))

			if !threshold.Mode.Null {
				fieldConfig.Thresholds.Mode = threshold.Mode.Value
			}

			for i, step := range threshold.Steps {
				s := grafana.ThresholdStep{
					Color: step.Color.Value,
				}

				if !step.Value.Null {
					value := step.Value.Value
					s.Value = &value
				}

				steps[i] = s
			}

			if len(steps) > 0 {
				fieldConfig.Thresholds.Steps = steps
			}
		}
	}

	return fieldConfig
}

// updaters
func updateTextSize(options *grafana.TextSize, opts []TextSizeOptions) {
	for _, textSize := range opts {
		if !textSize.Title.Null {
			size := int(textSize.Title.Value)
			options.TitleSize = &size
		}

		if !textSize.Value.Null {
			size := int(textSize.Value.Value)
			options.ValueSize = &size
		}
	}
}

func updateReduceOptions(options *grafana.ReduceOptions, opts []ReduceOptions) {
	for _, reducer := range opts {
		if !reducer.Values.Null {
			options.Values = reducer.Values.Value
		}

		if !reducer.Fields.Null {
			options.Fields = reducer.Fields.Value
		}

		if !reducer.Limit.Null {
			limit := int(reducer.Limit.Value)
			options.Limit = &limit
		}

		if !reducer.Calculation.Null {
			options.Calcs = []string{reducer.Calculation.Value}
		}
	}
}

// etc
func hashcode(s []byte) int {
	v := int(crc32.ChecksumIEEE(s))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}
