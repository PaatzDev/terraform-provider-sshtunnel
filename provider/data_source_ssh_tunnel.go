package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSshTunnel() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSshTunnelRead,
		Schema: map[string]*schema.Schema{
			"local_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceSshTunnelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tunnel := m.(*SSHTunnel)

	if err := d.Set("local_host", tunnel.Local.Host); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("local_port", tunnel.Local.Port); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
