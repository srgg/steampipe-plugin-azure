package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/datafactory/mgmt/2018-06-01/datafactory"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableAzureDataFactoryDataset(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_data_factory_dataset",
		Description: "Azure Data Factory Dataset",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.AllColumns([]string{"name", "resource_group", "factory_name"}),
			Hydrate:           getDataFactoryDataset,
			ShouldIgnoreError: isNotFoundError([]string{"ResourceNotFound", "ResourceGroupNotFound", "404"}),
		},
		List: &plugin.ListConfig{
			Hydrate:       listDataFactoryDatasets,
			ParentHydrate: listDataFactories,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The resource name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The resource identifier.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "factory_name",
				Description: "Name of the factory the dataset belongs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "etag",
				Description: "An unique read-only string that changes whenever the resource is updated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The resource type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "properties",
				Description: "Dataset properties.",
				Type:        proto.ColumnType_JSON,
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ID").Transform(idToAkas),
			},

			// Azure standard column
			{
				Name:        "resource_group",
				Description: ColumnDescriptionResourceGroup,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID").Transform(extractResourceGroupFromID),
			},
			{
				Name:        "subscription_id",
				Description: ColumnDescriptionSubscription,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID").Transform(idToSubscriptionID),
			},
		},
	}
}

type DatasetInfo = struct {
	datafactory.DatasetResource
	FactoryName string
}

//// LIST FUNCTION

func listDataFactoryDatasets(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	// Get factory details
	factoryInfo := h.Item.(datafactory.Factory)
	resourceGroup := strings.Split(*factoryInfo.ID, "/")[4]

	datasetClient := datafactory.NewDatasetsClient(subscriptionID)
	datasetClient.Authorizer = session.Authorizer

	result, err := datasetClient.ListByFactory(ctx, resourceGroup, *factoryInfo.Name)
	if err != nil {
		return nil, err
	}
	for _, dataset := range result.Values() {
		d.StreamListItem(ctx, DatasetInfo{dataset, *factoryInfo.Name})
	}

	for result.NotDone() {
		err = result.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, dataset := range result.Values() {
			d.StreamListItem(ctx, DatasetInfo{dataset, *factoryInfo.Name})
		}
	}
	return nil, err
}

//// HYDRATE FUNCTIONS

func getDataFactoryDataset(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getDataFactoryDataset")

	// Create session
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	datasetClient := datafactory.NewDatasetsClient(subscriptionID)
	datasetClient.Authorizer = session.Authorizer

	datasetName := d.KeyColumnQuals["name"].GetStringValue()
	resourceGroup := d.KeyColumnQuals["resource_group"].GetStringValue()
	factoryName := d.KeyColumnQuals["factory_name"].GetStringValue()

	// Return nil, of no input provided
	if datasetName == "" || resourceGroup == "" || factoryName == "" {
		return nil, nil
	}

	op, err := datasetClient.Get(ctx, resourceGroup, factoryName, datasetName, "")
	if err != nil {
		return nil, err
	}

	// In some cases resource does not give any notFound error
	// instead of notFound error, it returns empty data
	if op.ID != nil {
		return DatasetInfo{op, factoryName}, nil
	}

	return nil, nil
}
