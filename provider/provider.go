package provider

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"golang.org/x/crypto/ssh"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"gateway_host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The hostname or ip address of the remote ssh jump box server.",
			},
			"gateway_port": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The port of the remote sh jump box.",
			},
			"remote_host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The hostname or ip address of the target server. (I.e. database host)",
			},
			"remote_port": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The port of the target server. (I.e. database host)",
			},
			"gateway_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The username of the remote jump boy ssh server.",
			},
			"gateway_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The password for the remote jump box ssh server (Do not set this if you want to use private key authentication)",
			},
			"private_key_pem_string": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The raw pem encoded string of the private key (Do not set this if you want to use password authentication)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"ssh_tunnel": dataSourceSshTunnel(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	gateway_host := d.Get("gateway_host").(string)
	gateway_port := d.Get("gateway_port").(int)
	remote_host := d.Get("remote_host").(string)
	remote_port := d.Get("remote_port").(int)
	gateway_username := d.Get("").(string)
	gateway_password := d.Get("").(string)
	private_key_pem_string := d.Get("private_key_pem_string").(string)

	if (gateway_host != "") && (remote_host != "") && (gateway_username != "") {

		if gateway_port == -1 {
			gateway_port = 22
		}
		if remote_port == -1 {
			remote_port = 22
		}

		var auth ssh.AuthMethod

		auth = nil

		if gateway_password != "" {
			auth = ssh.Password(gateway_password)
		} else if private_key_pem_string != "" {
			auth = PrivateKey(private_key_pem_string)
		}

		sshTunnel := makeSshTunnel(gateway_host, gateway_port, remote_host, remote_port, gateway_username, auth)

		return sshTunnel, diags
	}

	return nil, diags
}

func makeSshTunnel(gatewayHost string, gatewayPort int, remoteHost string, remotePort int, gatewayUser string, auth ssh.AuthMethod) *SSHTunnel {
	tunnel := NewSSHTunnel(gatewayHost, gatewayPort, gatewayUser, auth, remoteHost, remotePort)

	tunnel.Log = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)

	go tunnel.Start()

	time.Sleep(100 * time.Millisecond)

	return tunnel
}
