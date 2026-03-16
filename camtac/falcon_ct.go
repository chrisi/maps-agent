package camtac

import "encoding/xml"

type CTRecords struct {
	XMLName xml.Name `xml:"CTRecords" json:"-"`
	CTs     []CT     `xml:"CT" json:"cts"`
}

type CT struct {
	Num                    int     `xml:"Num,attr" json:"num"`
	Id                     int     `xml:"Id" json:"id"`
	CollisionType          int     `xml:"CollisionType" json:"collisionType"`
	CollisionRadius        float64 `xml:"CollisionRadius" json:"collisionRadius"`
	Domain                 int     `xml:"Domain" json:"domain"`
	Class                  int     `xml:"Class" json:"class"`
	Type                   int     `xml:"Type" json:"type"`
	SubType                int     `xml:"SubType" json:"subType"`
	Specific               int     `xml:"Specific" json:"specific"`
	Owner                  int     `xml:"Owner" json:"owner"`
	Class6                 int     `xml:"Class_6" json:"class6"`
	Class7                 int     `xml:"Class_7" json:"class7"`
	UpdateRate             int     `xml:"UpdateRate" json:"updateRate"`
	UpdateTolerance        int     `xml:"UpdateTolerance" json:"updateTolerance"`
	FineUpdateRange        float64 `xml:"FineUpdateRange" json:"fineUpdateRange"`
	FineUpdateForceRange   float64 `xml:"FineUpdateForceRange" json:"fineUpdateForceRange"`
	FineUpdateMultiplier   float64 `xml:"FineUpdateMultiplier" json:"fineUpdateMultiplier"`
	DamageSeed             int     `xml:"DamageSeed" json:"damageSeed"`
	HitPoints              int     `xml:"HitPoints" json:"hitPoints"`
	MajorRev               int     `xml:"MajorRev" json:"majorRev"`
	MinRev                 int     `xml:"MinRev" json:"minRev"`
	CreatePriority         int     `xml:"CreatePriority" json:"createPriority"`
	ManagementDomain       int     `xml:"ManagementDomain" json:"managementDomain"`
	Transferable           int     `xml:"Transferable" json:"transferable"`
	Private                int     `xml:"Private" json:"private"`
	Tangible               int     `xml:"Tangible" json:"tangible"`
	Collidable             int     `xml:"Collidable" json:"collidable"`
	Global                 int     `xml:"Global" json:"global"`
	Persistent             int     `xml:"Persistent" json:"persistent"`
	GraphicsNormal         int     `xml:"GraphicsNormal" json:"graphicsNormal"`
	GraphicsRepaired       int     `xml:"GraphicsRepaired" json:"graphicsRepaired"`
	GraphicsDamaged        int     `xml:"GraphicsDamaged" json:"graphicsDamaged"`
	GraphicsDestroyed      int     `xml:"GraphicsDestroyed" json:"graphicsDestroyed"`
	GraphicsLeftDestroyed  int     `xml:"GraphicsLeftDestroyed" json:"graphicsLeftDestroyed"`
	GraphicsRightDestroyed int     `xml:"GraphicsRightDestroyed" json:"graphicsRightDestroyed"`
	GraphicsBothDestroyed  int     `xml:"GraphicsBothDestroyed" json:"graphicsBothDestroyed"`
	MoverDefinitionData    int     `xml:"MoverDefinitionData" json:"moverDefinitionData"`
	EntityType             int     `xml:"EntityType" json:"entityType"`
	EntityIdx              int     `xml:"EntityIdx" json:"entityIdx"`
}
