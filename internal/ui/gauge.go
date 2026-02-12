package ui

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Gauge is a speedometer-style gauge widget
type Gauge struct {
	widget.BaseWidget
	value    float64
	maxValue float64
	label    string
	unit     string
}

// NewGauge creates a new gauge widget
func NewGauge(label string, value, maxValue float64, unit string) *Gauge {
	g := &Gauge{
		value:    value,
		maxValue: maxValue,
		label:    label,
		unit:     unit,
	}
	g.ExtendBaseWidget(g)
	return g
}

// SetValue updates the gauge value
func (g *Gauge) SetValue(value float64) {
	g.value = value
	g.Refresh()
}

// CreateRenderer implements fyne.Widget
func (g *Gauge) CreateRenderer() fyne.WidgetRenderer {
	return &gaugeRenderer{gauge: g}
}

type gaugeRenderer struct {
	gauge *Gauge
}

func (r *gaugeRenderer) Layout(size fyne.Size) {}

func (r *gaugeRenderer) MinSize() fyne.Size {
	return fyne.NewSize(120, 100)
}

func (r *gaugeRenderer) Refresh() {}

func (r *gaugeRenderer) Destroy() {}

func (r *gaugeRenderer) Objects() []fyne.CanvasObject {
	return r.createGaugeObjects()
}

func (r *gaugeRenderer) createGaugeObjects() []fyne.CanvasObject {
	g := r.gauge
	objects := []fyne.CanvasObject{}

	// Calculate percentage
	percentage := 0.0
	if g.maxValue > 0 {
		percentage = (g.value / g.maxValue) * 100
	}

	// Determine color based on value
	var gaugeColor color.Color
	if percentage < 50 {
		gaugeColor = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green
	} else if percentage < 80 {
		gaugeColor = color.NRGBA{R: 255, G: 193, B: 7, A: 255} // Yellow/Orange
	} else {
		gaugeColor = color.NRGBA{R: 244, G: 67, B: 54, A: 255} // Red
	}

	// Background arc (gray)
	bgArc := canvas.NewCircle(color.NRGBA{R: 60, G: 50, B: 70, A: 255})
	bgArc.StrokeWidth = 8
	bgArc.StrokeColor = color.NRGBA{R: 60, G: 50, B: 70, A: 255}
	bgArc.Resize(fyne.NewSize(100, 100))
	objects = append(objects, bgArc)

	// Value arc
	valueArc := canvas.NewCircle(color.Transparent)
	valueArc.StrokeWidth = 8
	valueArc.StrokeColor = gaugeColor
	valueArc.Resize(fyne.NewSize(100, 100))
	objects = append(objects, valueArc)

	// Value text
	valueText := canvas.NewText(fmt.Sprintf("%.1f%s", g.value, g.unit), PurpleText)
	valueText.TextSize = 18
	valueText.TextStyle = fyne.TextStyle{Bold: true}
	valueText.Alignment = fyne.TextAlignCenter
	objects = append(objects, valueText)

	// Label text
	labelText := canvas.NewText(g.label, PurpleSecondary)
	labelText.TextSize = 12
	labelText.Alignment = fyne.TextAlignCenter
	objects = append(objects, labelText)

	return objects
}

// CreateGaugeDisplay creates a visual gauge display using standard Fyne widgets
func CreateGaugeDisplay(label string, value, maxValue float64, unit string) fyne.CanvasObject {
	percentage := 0.0
	if maxValue > 0 {
		percentage = (value / maxValue) * 100
	}

	// Determine color based on value
	var gaugeColor color.Color
	if percentage < 50 {
		gaugeColor = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green
	} else if percentage < 80 {
		gaugeColor = color.NRGBA{R: 255, G: 193, B: 7, A: 255} // Yellow
	} else {
		gaugeColor = color.NRGBA{R: 244, G: 67, B: 54, A: 255} // Red
	}

	// Create arc segments to simulate a gauge
	arcContainer := createArcGauge(percentage, gaugeColor)

	// Value label
	valueLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("%.1f%s", value, unit),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Title label
	titleLabel := widget.NewLabelWithStyle(
		label,
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	return container.NewVBox(
		arcContainer,
		valueLabel,
		titleLabel,
	)
}

// createArcGauge creates a visual arc gauge using rectangles
func createArcGauge(percentage float64, gaugeColor color.Color) fyne.CanvasObject {
	const (
		width      = 120
		height     = 60
		segments   = 20
		startAngle = 180.0 // Start from left
		endAngle   = 0.0   // End at right (180 degree arc)
	)

	objects := []fyne.CanvasObject{}

	// Background arc
	bgColor := color.NRGBA{R: 50, G: 40, B: 60, A: 255}

	// Calculate how many segments to fill
	filledSegments := int(math.Round(float64(segments) * percentage / 100.0))

	centerX := float32(width / 2)
	centerY := float32(height)
	radius := float32(50)

	for i := 0; i < segments; i++ {
		// Calculate angle for this segment
		angle := startAngle - (float64(i) * (180.0 / float64(segments-1)))
		rad := angle * math.Pi / 180.0

		// Position
		x := centerX + radius*float32(math.Cos(rad)) - 4
		y := centerY - radius*float32(math.Sin(rad)) - 4

		// Create segment
		var segColor color.Color
		if i < filledSegments {
			segColor = gaugeColor
		} else {
			segColor = bgColor
		}

		segment := canvas.NewCircle(segColor)
		segment.Resize(fyne.NewSize(8, 8))
		segment.Move(fyne.NewPos(x, y))
		objects = append(objects, segment)
	}

	return container.NewWithoutLayout(objects...)
}

// CreateCompactGauge creates a compact gauge with progress bar style
func CreateCompactGauge(label string, value float64, unit string) fyne.CanvasObject {
	percentage := value
	if percentage > 100 {
		percentage = 100
	}

	// Determine color based on value
	var barColor color.Color
	if percentage < 50 {
		barColor = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green
	} else if percentage < 80 {
		barColor = color.NRGBA{R: 255, G: 193, B: 7, A: 255} // Yellow
	} else {
		barColor = color.NRGBA{R: 244, G: 67, B: 54, A: 255} // Red
	}

	// Background bar
	bgBar := canvas.NewRectangle(color.NRGBA{R: 50, G: 40, B: 60, A: 255})
	bgBar.Resize(fyne.NewSize(200, 20))
	bgBar.CornerRadius = 10

	// Value bar
	valueBar := canvas.NewRectangle(barColor)
	valueBar.Resize(fyne.NewSize(float32(percentage*2), 20))
	valueBar.CornerRadius = 10

	barStack := container.NewStack(bgBar, container.NewWithoutLayout(valueBar))

	// Label
	titleLabel := widget.NewLabel(label)

	// Value text
	valueText := widget.NewLabelWithStyle(
		fmt.Sprintf("%.1f%s", value, unit),
		fyne.TextAlignTrailing,
		fyne.TextStyle{Bold: true},
	)

	topRow := container.NewBorder(nil, nil, titleLabel, valueText)

	return container.NewVBox(topRow, barStack)
}

// CreateSpeedometerGauge creates a semicircle speedometer gauge
func CreateSpeedometerGauge(label string, value float64, unit string) fyne.CanvasObject {
	percentage := value
	if percentage > 100 {
		percentage = 100
	}

	// Determine color based on value
	var arcColor color.Color
	if percentage < 50 {
		arcColor = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green
	} else if percentage < 80 {
		arcColor = color.NRGBA{R: 255, G: 193, B: 7, A: 255} // Yellow
	} else {
		arcColor = color.NRGBA{R: 244, G: 67, B: 54, A: 255} // Red
	}

	// Create the arc visualization
	arcDisplay := createSemicircleGauge(percentage, arcColor)

	// Value in center
	valueLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("%.1f%s", value, unit),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Label below
	titleLabel := widget.NewLabelWithStyle(
		label,
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	return container.NewVBox(
		container.NewCenter(arcDisplay),
		valueLabel,
		titleLabel,
	)
}

func createSemicircleGauge(percentage float64, arcColor color.Color) fyne.CanvasObject {
	const (
		width    = 140
		height   = 80
		segments = 30
	)

	objects := []fyne.CanvasObject{}

	bgColor := color.NRGBA{R: 50, G: 40, B: 60, A: 255}

	filledSegments := int(math.Round(float64(segments) * percentage / 100.0))

	centerX := float32(width / 2)
	centerY := float32(height - 10)
	radius := float32(55)

	for i := 0; i < segments; i++ {
		// Calculate angle (180 to 0 degrees)
		angle := 180.0 - (float64(i) * (180.0 / float64(segments-1)))
		rad := angle * math.Pi / 180.0

		x := centerX + radius*float32(math.Cos(rad)) - 5
		y := centerY - radius*float32(math.Sin(rad)) - 5

		var segColor color.Color
		if i < filledSegments {
			// Gradient effect - interpolate color
			segColor = arcColor
		} else {
			segColor = bgColor
		}

		segment := canvas.NewCircle(segColor)
		segment.Resize(fyne.NewSize(10, 10))
		segment.Move(fyne.NewPos(x, y))
		objects = append(objects, segment)
	}

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(width, height))
	return cont
}
