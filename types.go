package main

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Objective struct {
	Name   string `json:"name"`
	Owner  int    `json:"owner"`
	OcdIdx int    `json:"ocdIdx"`
	Type   int    `json:"type"`
	Pos    Point  `json:"pos"`
}
