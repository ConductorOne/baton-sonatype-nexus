package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// The user resource type is for all user objects from the database.
var userResourceType = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

var roleResourceType = &v2.ResourceType{
	Id:          "role",
	DisplayName: "Role",
	Description: "A role in Nexus",
	Traits: []v2.ResourceType_Trait{
		v2.ResourceType_TRAIT_ROLE,
	},
}
