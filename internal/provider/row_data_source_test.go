package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRowDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccRowDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdashboard_row.test", "title", "Test"),
					resource.TestCheckResourceAttr("data.gdashboard_row.test", "json", testAccRowDataSourceExpectedJson),
				),
			},
			{
				Config: testAccRowDataSourceProviderDefaultsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdashboard_row.test", "title", "Test"),
					resource.TestCheckResourceAttr("data.gdashboard_row.test", "json", testAccRowDataSourceProviderDefaultsConfigExpectedJson),
				),
			},
		},
	})
}

const testAccRowDataSourceConfig = `
data "gdashboard_row" "test" {
  title = "Test"

  graph {
    collapsed = true
  }
}
`

const testAccRowDataSourceExpectedJson = `{
  "editable": false,
  "error": false,
  "gridPos": {},
  "id": 0,
  "isNew": true,
  "span": 12,
  "title": "Test",
  "transparent": false,
  "type": "row",
  "panels": null,
  "collapsed": true
}`

const testAccRowDataSourceProviderDefaultsConfig = `
data "gdashboard_row" "test" {
  title = "Test"
}
`

const testAccRowDataSourceProviderDefaultsConfigExpectedJson = `{
  "editable": false,
  "error": false,
  "gridPos": {},
  "id": 0,
  "isNew": true,
  "span": 12,
  "title": "Test",
  "transparent": false,
  "type": "row",
  "panels": null,
  "collapsed": false
}`
