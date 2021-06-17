package cosmos

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-01-15/documentdb"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/cosmos/common"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/cosmos/migration"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/cosmos/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/cosmos/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceCosmosDbCassandraKeyspace() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceCosmosDbCassandraKeyspaceCreate,
		Read:   resourceCosmosDbCassandraKeyspaceRead,
		Update: resourceCosmosDbCassandraKeyspaceUpdate,
		Delete: resourceCosmosDbCassandraKeyspaceDelete,

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.CassandraKeyspaceV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.CosmosEntityName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.CosmosAccountName,
			},

			"throughput": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.CosmosThroughput,
			},

			"autoscale_settings": common.DatabaseAutoscaleSettingsSchema(),
		},
	}
}

func resourceCosmosDbCassandraKeyspaceCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.CassandraClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	account := d.Get("account_name").(string)

	existing, err := client.GetCassandraKeyspace(ctx, resourceGroup, account, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("Error checking for presence of creating Cosmos Cassandra Keyspace %q (Account: %q): %+v", name, account, err)
		}
	} else {
		if existing.ID == nil && *existing.ID == "" {
			return fmt.Errorf("Error generating import ID for Cosmos Cassandra Keyspace %q (Account: %q)", name, account)
		}

		return tf.ImportAsExistsError("azurerm_cosmosdb_cassandra_keyspace", *existing.ID)
	}

	db := documentdb.CassandraKeyspaceCreateUpdateParameters{
		CassandraKeyspaceCreateUpdateProperties: &documentdb.CassandraKeyspaceCreateUpdateProperties{
			Resource: &documentdb.CassandraKeyspaceResource{
				ID: &name,
			},
			Options: &documentdb.CreateUpdateOptions{},
		},
	}

	if throughput, hasThroughput := d.GetOk("throughput"); hasThroughput {
		if throughput != 0 {
			db.CassandraKeyspaceCreateUpdateProperties.Options.Throughput = common.ConvertThroughputFromResourceData(throughput)
		}
	}

	if _, hasAutoscaleSettings := d.GetOk("autoscale_settings"); hasAutoscaleSettings {
		db.CassandraKeyspaceCreateUpdateProperties.Options.AutoscaleSettings = common.ExpandCosmosDbAutoscaleSettings(d)
	}

	future, err := client.CreateUpdateCassandraKeyspace(ctx, resourceGroup, account, name, db)
	if err != nil {
		return fmt.Errorf("Error issuing create/update request for Cosmos Cassandra Keyspace %q (Account: %q): %+v", name, account, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting on create/update future for Cosmos Cassandra Keyspace %q (Account: %q): %+v", name, account, err)
	}

	resp, err := client.GetCassandraKeyspace(ctx, resourceGroup, account, name)
	if err != nil {
		return fmt.Errorf("Error making get request for Cosmos Cassandra Keyspace %q (Account: %q): %+v", name, account, err)
	}

	if resp.ID == nil {
		return fmt.Errorf("Error getting ID from Cosmos Cassandra Keyspace %q (Account: %q)", name, account)
	}

	d.SetId(*resp.ID)

	return resourceCosmosDbCassandraKeyspaceRead(d, meta)
}

func resourceCosmosDbCassandraKeyspaceUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.CassandraClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CassandraKeyspaceID(d.Id())
	if err != nil {
		return err
	}

	err = common.CheckForChangeFromAutoscaleAndManualThroughput(d)
	if err != nil {
		return fmt.Errorf("Error updating Cosmos Cassandra Keyspace %q (Account: %q) - %+v", id.Name, id.DatabaseAccountName, err)
	}

	db := documentdb.CassandraKeyspaceCreateUpdateParameters{
		CassandraKeyspaceCreateUpdateProperties: &documentdb.CassandraKeyspaceCreateUpdateProperties{
			Resource: &documentdb.CassandraKeyspaceResource{
				ID: &id.Name,
			},
			Options: &documentdb.CreateUpdateOptions{},
		},
	}

	future, err := client.CreateUpdateCassandraKeyspace(ctx, id.ResourceGroup, id.DatabaseAccountName, id.Name, db)
	if err != nil {
		return fmt.Errorf("Error issuing create/update request for Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.ResourceGroup, id.DatabaseAccountName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting on create/update future for Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.ResourceGroup, id.DatabaseAccountName, err)
	}

	if common.HasThroughputChange(d) {
		throughputParameters := common.ExpandCosmosDBThroughputSettingsUpdateParameters(d)
		throughputFuture, err := client.UpdateCassandraKeyspaceThroughput(ctx, id.ResourceGroup, id.DatabaseAccountName, id.Name, *throughputParameters)
		if err != nil {
			if response.WasNotFound(throughputFuture.Response()) {
				return fmt.Errorf("Error setting Throughput for Cosmos Cassandra Keyspace %q (Account: %q): %+v - "+
					"If the collection has not been created with an initial throughput, you cannot configure it later.", id.Name, id.DatabaseAccountName, err)
			}
		}

		if err = throughputFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("Error waiting on ThroughputUpdate future for Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.Name, id.DatabaseAccountName, err)
		}
	}

	return resourceCosmosDbCassandraKeyspaceRead(d, meta)
}

func resourceCosmosDbCassandraKeyspaceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.CassandraClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CassandraKeyspaceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetCassandraKeyspace(ctx, id.ResourceGroup, id.DatabaseAccountName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading Cosmos Cassandra Keyspace %q (Account: %q) - removing from state", id.Name, id.DatabaseAccountName)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.Name, id.DatabaseAccountName, err)
	}

	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("account_name", id.DatabaseAccountName)
	if props := resp.CassandraKeyspaceGetProperties; props != nil {
		if res := props.Resource; res != nil {
			d.Set("name", res.ID)
		}
	}

	throughputResp, err := client.GetCassandraKeyspaceThroughput(ctx, id.ResourceGroup, id.DatabaseAccountName, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(throughputResp.Response) {
			return fmt.Errorf("Error reading Throughput on Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.Name, id.DatabaseAccountName, err)
		} else {
			d.Set("throughput", nil)
			d.Set("autoscale_settings", nil)
		}
	} else {
		common.SetResourceDataThroughputFromResponse(throughputResp, d)
	}

	return nil
}

func resourceCosmosDbCassandraKeyspaceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.CassandraClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CassandraKeyspaceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.DeleteCassandraKeyspace(ctx, id.ResourceGroup, id.DatabaseAccountName, id.Name)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error deleting Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.Name, id.DatabaseAccountName, err)
		}
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return fmt.Errorf("Error waiting on delete future for Cosmos Cassandra Keyspace %q (Account: %q): %+v", id.Name, id.DatabaseAccountName, err)
	}

	return nil
}
