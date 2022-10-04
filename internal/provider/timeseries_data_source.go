package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/iRevive/terraform-provider-gdashboard/internal/provider/grafana"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &TimeseriesDataSource{}

func NewTimeseriesDataSource() datasource.DataSource {
	return &TimeseriesDataSource{}
}

// TimeseriesDataSource defines the data source implementation.
type TimeseriesDataSource struct {
	Defaults TimeseriesDefaults
}

type TimeseriesDefaults struct {
	Legend  TimeseriesLegendDefault
	Tooltip TimeseriesTooltipDefaults
	Field   FieldDefaults
	Axis    AxisDefaults
	Graph   TimeseriesGraphDefault
}

type TimeseriesGraphDefault struct {
	DrawStyle         string
	LineInterpolation string
	LineWidth         int
	FillOpacity       int
	GradientMode      string
	LineStyle         string
	SpanNulls         bool
	ShowPoints        string
	PointSize         int
	StackSeries       string
}

type TimeseriesTooltipDefaults struct {
	Mode string
}

type TimeseriesLegendDefault struct {
	Calculations []string
	DisplayMode  string
	Placement    string
}

// TimeseriesDataSourceModel describes the data source data model.
type TimeseriesDataSourceModel struct {
	Id      types.String               `tfsdk:"id"`
	Json    types.String               `tfsdk:"json"`
	Title   types.String               `tfsdk:"title"`
	Targets []Target                   `tfsdk:"targets"`
	Legend  []TimeseriesLegendOptions  `tfsdk:"legend"`
	Tooltip []TimeseriesTooltipOptions `tfsdk:"tooltip"`
	Field   []FieldOptions             `tfsdk:"field"`
	Axis    []AxisOptions              `tfsdk:"axis"`
	Graph   []TimeseriesGraphOptions   `tfsdk:"graph"`
}

type TimeseriesLegendOptions struct {
	Calculations []types.String `tfsdk:"calculations"`
	DisplayMode  types.String   `tfsdk:"display_mode"`
	Placement    types.String   `tfsdk:"placement"`
}

type TimeseriesTooltipOptions struct {
	Mode types.String `tfsdk:"mode"`
}

type TimeseriesGraphOptions struct {
	DrawStyle         types.String `tfsdk:"draw_style"`
	LineInterpolation types.String `tfsdk:"line_interpolation"`
	LineWidth         types.Int64  `tfsdk:"line_width"`
	FillOpacity       types.Int64  `tfsdk:"fill_opacity"`
	GradientMode      types.String `tfsdk:"gradient_mode"`
	LineStyle         types.String `tfsdk:"line_style"`
	SpanNulls         types.Bool   `tfsdk:"span_nulls"`
	ShowPoints        types.String `tfsdk:"show_points"`
	PointSize         types.Int64  `tfsdk:"point_size"`
	StackSeries       types.String `tfsdk:"stack_series"`
}

func (d *TimeseriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_timeseries"
}

func timeseriesTooltipBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Attributes: map[string]tfsdk.Attribute{
			"mode": {
				Type:        types.StringType,
				Required:    true,
				Description: "When you hover your cursor over the visualization, Grafana can display tooltips",
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("multi", "single", "hidden"),
				},
			},
		},
	}
}

func timeseriesGraphBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Attributes: map[string]tfsdk.Attribute{
			"draw_style": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("line", "bars", "points"),
				},
			},
			"line_interpolation": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("linear", "smooth", "stepBefore", "stepAfter"),
				},
			},
			"line_width": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"fill_opacity": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"gradient_mode": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("none", "opacity", "hue", "scheme"),
				},
			},
			"line_style": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("solid", "dash", "dots"),
				},
			},
			"span_nulls": {
				Type:     types.BoolType,
				Optional: true,
			},
			"show_points": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("auto", "never", "always"),
				},
			},
			"point_size": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"stack_series": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("none", "normal", "percent"),
				},
			},
		},
	}
}

func timeseriesLegendBlock() tfsdk.Block {
	return tfsdk.Block{
		NestingMode: tfsdk.BlockNestingModeList,
		MinItems:    0,
		MaxItems:    1,
		Attributes: map[string]tfsdk.Attribute{
			"calculations": {
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
				Description: "Choose which of the standard calculations to show in the legend: min, max, mean",
			},
			"display_mode": {
				Type:        types.StringType,
				Optional:    true,
				Description: "Use these settings to define how the legend appears in your visualization: list, table, hidden",
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("list", "table", "hidden"),
				},
			},
			"placement": {
				Type:        types.StringType,
				Optional:    true,
				Description: "Choose where to display the legend: bottom, right",
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf("bottom", "right"),
				},
			},
		},
	}
}

func (d *TimeseriesDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Timeseries panel data source",

		Blocks: map[string]tfsdk.Block{
			"targets": targetBlock(),
			"legend":  timeseriesLegendBlock(),
			"tooltip": timeseriesTooltipBlock(),
			"field":   fieldBlock(),
			"axis":    axisBlock(),
			"graph":   timeseriesGraphBlock(),
		},

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"json": {
				Type:     types.StringType,
				Computed: true,
			},
			"title": {
				Type:        types.StringType,
				Required:    true,
				Description: "The title of the panel",
			},
		},
	}, nil
}

func (d *TimeseriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	defaults, ok := req.ProviderData.(Defaults)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Defaults, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}

	d.Defaults = defaults.Timeseries
}

func (d *TimeseriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TimeseriesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	targets := createTargets(data.Targets)

	legendOptions := grafana.TimeseriesLegendOptions{
		Calcs:       d.Defaults.Legend.Calculations,
		DisplayMode: d.Defaults.Legend.DisplayMode,
		Placement:   d.Defaults.Legend.Placement,
	}

	tooltipOptions := grafana.TimeseriesTooltipOptions{
		Mode: d.Defaults.Tooltip.Mode,
	}

	for _, legend := range data.Legend {
		if len(legend.Calculations) > 0 {
			calculations := make([]string, len(legend.Calculations))
			for i, calc := range legend.Calculations {
				calculations[i] = calc.Value
			}

			legendOptions.Calcs = calculations
		}

		if !legend.DisplayMode.Null {
			legendOptions.DisplayMode = legend.DisplayMode.Value
		}

		if !legend.Placement.Null {
			legendOptions.Placement = legend.Placement.Value
		}
	}

	for _, tooltip := range data.Tooltip {
		tooltipOptions.Mode = tooltip.Mode.Value
	}

	fieldConfig := createFieldConfig(d.Defaults.Field, data.Field)

	fieldConfig.Custom = grafana.FieldConfigCustom{
		// graph
		DrawStyle:         d.Defaults.Graph.DrawStyle,
		LineInterpolation: d.Defaults.Graph.LineInterpolation,
		LineWidth:         d.Defaults.Graph.LineWidth,
		FillOpacity:       d.Defaults.Graph.FillOpacity,
		GradientMode:      d.Defaults.Graph.GradientMode,
		SpanNulls:         d.Defaults.Graph.SpanNulls,
		ShowPoints:        d.Defaults.Graph.ShowPoints,
		PointSize:         d.Defaults.Graph.PointSize,
		// axis
		AxisLabel:     d.Defaults.Axis.Label,
		AxisPlacement: d.Defaults.Axis.Placement,
		AxisSoftMin:   d.Defaults.Axis.SoftMin,
		AxisSoftMax:   d.Defaults.Axis.SoftMax,
	}

	fieldConfig.Custom.LineStyle.Fill = d.Defaults.Graph.LineStyle
	fieldConfig.Custom.Stacking.Mode = d.Defaults.Graph.StackSeries
	fieldConfig.Custom.ScaleDistribution.Type = d.Defaults.Axis.Scale.Type
	fieldConfig.Custom.ScaleDistribution.Log = d.Defaults.Axis.Scale.Log

	for _, axis := range data.Axis {
		if !axis.Label.Null {
			fieldConfig.Custom.AxisLabel = axis.Label.Value
		}

		if !axis.Placement.Null {
			fieldConfig.Custom.AxisPlacement = axis.Placement.Value
		}

		if !axis.SoftMin.Null {
			min := int(axis.SoftMin.Value)
			fieldConfig.Custom.AxisSoftMin = &min
		}

		if !axis.SoftMax.Null {
			max := int(axis.SoftMax.Value)
			fieldConfig.Custom.AxisSoftMax = &max
		}

		for _, scale := range axis.Scale {
			if !scale.Type.Null {
				fieldConfig.Custom.ScaleDistribution.Type = scale.Type.Value
			}

			if !scale.Log.Null {
				fieldConfig.Custom.ScaleDistribution.Log = int(scale.Log.Value)
			}
		}
	}

	for _, graph := range data.Graph {
		if !graph.DrawStyle.Null {
			fieldConfig.Custom.DrawStyle = graph.DrawStyle.Value
		}

		if !graph.LineInterpolation.Null {
			fieldConfig.Custom.LineInterpolation = graph.LineInterpolation.Value
		}

		if !graph.LineWidth.Null {
			fieldConfig.Custom.LineWidth = int(graph.LineWidth.Value)
		}

		if !graph.FillOpacity.Null {
			fieldConfig.Custom.FillOpacity = int(graph.FillOpacity.Value)
		}

		if !graph.GradientMode.Null {
			fieldConfig.Custom.GradientMode = graph.GradientMode.Value
		}

		if !graph.LineStyle.Null {
			fieldConfig.Custom.LineStyle.Fill = graph.LineStyle.Value
		}

		if !graph.SpanNulls.Null {
			fieldConfig.Custom.SpanNulls = graph.SpanNulls.Value
		}

		if !graph.ShowPoints.Null {
			fieldConfig.Custom.ShowPoints = graph.ShowPoints.Value
		}

		if !graph.PointSize.Null {
			fieldConfig.Custom.PointSize = int(graph.PointSize.Value)
		}

		if !graph.StackSeries.Null {
			fieldConfig.Custom.Stacking.Mode = graph.StackSeries.Value
		}
	}

	panel := &grafana.Panel{
		CommonPanel: grafana.CommonPanel{
			OfType: grafana.TimeseriesType,
			Title:  data.Title.Value,
			Type:   "timeseries",
			Span:   12,
			IsNew:  true,
		},
		TimeseriesPanel: &grafana.TimeseriesPanel{
			Targets: targets,
			Options: grafana.TimeseriesOptions{
				Legend:  legendOptions,
				Tooltip: tooltipOptions,
			},
			FieldConfig: grafana.FieldConfig{
				Defaults: fieldConfig,
			},
		},
	}

	jsonData, err := json.MarshalIndent(panel, "", "  ")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not marshal json: %s", err))
		return
	}

	data.Json = types.String{Value: string(jsonData)}
	data.Id = types.String{Value: strconv.Itoa(hashcode(jsonData))}

	//resp.Diagnostics.AddError("Client Error", fmt.Sprintf("%s", string(jsonData)))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
