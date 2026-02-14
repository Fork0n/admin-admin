package ui

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ThresholdZone defines a color zone for the gauge arc
type ThresholdZone struct {
	Start float64     // Start value (percentage of max)
	End   float64     // End value (percentage of max)
	Color color.Color // Color for this zone
}

// GaugeConfig holds configuration for a Gauge widget
type GaugeConfig struct {
	Min        float64         // Minimum value (default: 0)
	Max        float64         // Maximum value (default: 100)
	StartAngle float64         // Start angle in degrees (default: -120)
	ArcSpan    float64         // Arc span in degrees (default: 240)
	Label      string          // Label text below gauge
	Unit       string          // Unit suffix for value display
	ArcColor   color.Color     // Default arc color
	Thresholds []ThresholdZone // Optional threshold color zones
}

// DefaultGaugeConfig returns default gauge configuration
func DefaultGaugeConfig() GaugeConfig {
	return GaugeConfig{
		Min:        0,
		Max:        100,
		StartAngle: 180, // Start from left
		ArcSpan:    180, // Go counter-clockwise to right
		Label:      "",
		Unit:       "%",
		ArcColor:   PurplePrimary,
		Thresholds: []ThresholdZone{
			{Start: 0, End: 50, Color: color.NRGBA{R: 138, G: 43, B: 226, A: 255}},  // Purple (Blue Violet)
			{Start: 50, End: 80, Color: color.NRGBA{R: 186, G: 85, B: 211, A: 255}}, // Medium Orchid
			{Start: 80, End: 100, Color: color.NRGBA{R: 255, G: 0, B: 255, A: 255}}, // Magenta/Bright purple
		},
	}
}

// Gauge is a radial gauge widget for displaying values
type Gauge struct {
	widget.BaseWidget

	mu           sync.RWMutex
	config       GaugeConfig
	currentValue float64 // Current animated value
	targetValue  float64 // Target value to animate towards

	// Animation control
	animating    bool
	animStop     chan struct{}
	animStopOnce sync.Once
}

// NewGauge creates a new Gauge widget with default configuration
func NewGauge(label string) *Gauge {
	config := DefaultGaugeConfig()
	config.Label = label
	return NewGaugeWithConfig(config)
}

// NewGaugeWithConfig creates a new Gauge widget with custom configuration
func NewGaugeWithConfig(config GaugeConfig) *Gauge {
	g := &Gauge{
		config:       config,
		currentValue: config.Min,
		targetValue:  config.Min,
		animStop:     make(chan struct{}),
	}
	g.ExtendBaseWidget(g)
	return g
}

// SetValue sets the target value with smooth animation
func (g *Gauge) SetValue(value float64) {
	g.mu.Lock()
	// Clamp value to min/max
	if value < g.config.Min {
		value = g.config.Min
	}
	if value > g.config.Max {
		value = g.config.Max
	}
	g.targetValue = value
	shouldStartAnim := !g.animating
	g.mu.Unlock()

	if shouldStartAnim {
		g.startAnimation()
	}
}

// GetValue returns the current displayed value
func (g *Gauge) GetValue() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentValue
}

// SetConfig updates the gauge configuration
func (g *Gauge) SetConfig(config GaugeConfig) {
	g.mu.Lock()
	g.config = config
	g.mu.Unlock()
	g.Refresh()
}

// startAnimation starts the animation loop
func (g *Gauge) startAnimation() {
	g.mu.Lock()
	if g.animating {
		g.mu.Unlock()
		return
	}
	g.animating = true
	g.mu.Unlock()

	go func() {
		ticker := time.NewTicker(33 * time.Millisecond) // ~30 FPS
		defer ticker.Stop()

		const epsilon = 0.01 // Small threshold for animation completion

		for {
			select {
			case <-g.animStop:
				g.mu.Lock()
				g.animating = false
				g.mu.Unlock()
				return
			case <-ticker.C:
				g.mu.Lock()
				diff := g.targetValue - g.currentValue

				// Check if animation is complete
				if math.Abs(diff) < epsilon {
					g.currentValue = g.targetValue
					g.animating = false
					g.mu.Unlock()

					// Final refresh on main thread
					g.safeRefresh()
					return
				}

				// Interpolate: smooth easing
				g.currentValue += diff * 0.15
				g.mu.Unlock()

				// Refresh on main thread
				g.safeRefresh()
			}
		}
	}()
}

// safeRefresh calls Refresh on the main UI thread
func (g *Gauge) safeRefresh() {
	if app := fyne.CurrentApp(); app != nil {
		if drv := app.Driver(); drv != nil {
			drv.DoFromGoroutine(func() {
				g.Refresh()
			}, false)
		}
	}
}

// StopAnimation stops the animation loop
func (g *Gauge) StopAnimation() {
	g.animStopOnce.Do(func() {
		close(g.animStop)
	})
}

// CreateRenderer implements fyne.Widget
func (g *Gauge) CreateRenderer() fyne.WidgetRenderer {
	g.mu.RLock()
	config := g.config
	g.mu.RUnlock()

	r := &gaugeRenderer{
		gauge:     g,
		arcBg:     make([]*canvas.Circle, 0),
		arcFg:     make([]*canvas.Circle, 0),
		needle:    canvas.NewLine(color.White),
		centerCap: canvas.NewCircle(PurpleSurface),
		valueText: canvas.NewText("0", PurpleText),
		labelText: canvas.NewText(config.Label, PurpleSecondary),
	}

	r.needle.StrokeWidth = 2
	r.valueText.TextStyle = fyne.TextStyle{Bold: true}
	r.valueText.Alignment = fyne.TextAlignCenter
	r.labelText.Alignment = fyne.TextAlignCenter

	r.initArcSegments()

	return r
}

// MinSize returns the minimum size of the gauge
func (g *Gauge) MinSize() fyne.Size {
	return fyne.NewSize(150, 150)
}

// gaugeRenderer handles the rendering of the Gauge widget
type gaugeRenderer struct {
	gauge *Gauge

	// Canvas objects (reused to avoid allocations)
	arcBg     []*canvas.Circle // Background arc segments
	arcFg     []*canvas.Circle // Foreground (filled) arc segments
	needle    *canvas.Line     // Needle line
	centerCap *canvas.Circle   // Center cap circle
	valueText *canvas.Text     // Value display text
	labelText *canvas.Text     // Label text

	// Cached values
	lastSize  fyne.Size
	lastValue float64
}

const arcSegments = 60 // Number of segments for smooth arc

// initArcSegments initializes the arc segment objects
func (r *gaugeRenderer) initArcSegments() {
	bgColor := color.NRGBA{R: 50, G: 40, B: 60, A: 255}

	for i := 0; i < arcSegments; i++ {
		// Background segment
		bgSeg := canvas.NewCircle(bgColor)
		r.arcBg = append(r.arcBg, bgSeg)

		// Foreground segment (initially transparent)
		fgSeg := canvas.NewCircle(color.Transparent)
		r.arcFg = append(r.arcFg, fgSeg)
	}
}

// Layout positions all elements within the given size
func (r *gaugeRenderer) Layout(size fyne.Size) {
	if size.Width < 10 || size.Height < 10 {
		return
	}

	r.gauge.mu.RLock()
	config := r.gauge.config
	currentValue := r.gauge.currentValue
	r.gauge.mu.RUnlock()

	// Calculate dimensions
	padding := float32(10)
	availableSize := fyne.NewSize(size.Width-padding*2, size.Height-padding*2)

	// Use the smaller dimension for the gauge
	gaugeSize := availableSize.Width
	if availableSize.Height < gaugeSize {
		gaugeSize = availableSize.Height
	}

	centerX := size.Width / 2
	centerY := size.Height/2 - 10 // Offset up to make room for label
	radius := gaugeSize * 0.4
	segmentSize := float32(math.Max(4, float64(gaugeSize)*0.06))

	// Calculate normalized value and filled segments
	normalized := (currentValue - config.Min) / (config.Max - config.Min)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	filledSegments := int(float64(arcSegments) * normalized)

	// Position arc segments
	for i := 0; i < arcSegments; i++ {
		// Calculate angle for this segment
		segmentNorm := float64(i) / float64(arcSegments-1)
		angle := config.StartAngle + segmentNorm*config.ArcSpan
		rad := angle * math.Pi / 180

		// Calculate position
		x := float32(float64(centerX) + float64(radius)*math.Cos(rad))
		y := float32(float64(centerY) + float64(radius)*math.Sin(rad))

		// Position and size background segment
		r.arcBg[i].Resize(fyne.NewSize(segmentSize, segmentSize))
		r.arcBg[i].Move(fyne.NewPos(x-segmentSize/2, y-segmentSize/2))

		// Position and size foreground segment
		r.arcFg[i].Resize(fyne.NewSize(segmentSize, segmentSize))
		r.arcFg[i].Move(fyne.NewPos(x-segmentSize/2, y-segmentSize/2))

		// Set foreground color based on fill and thresholds
		if i < filledSegments {
			segColor := r.getColorForValue(segmentNorm*100, config)
			r.arcFg[i].FillColor = segColor
		} else {
			r.arcFg[i].FillColor = color.Transparent
		}
	}

	// Position needle
	needleAngle := config.StartAngle + normalized*config.ArcSpan
	needleRad := needleAngle * math.Pi / 180
	needleLength := radius * 0.85
	needleEndX := float32(float64(centerX) + float64(needleLength)*math.Cos(needleRad))
	needleEndY := float32(float64(centerY) + float64(needleLength)*math.Sin(needleRad))

	r.needle.Position1 = fyne.NewPos(centerX, centerY)
	r.needle.Position2 = fyne.NewPos(needleEndX, needleEndY)
	r.needle.StrokeWidth = float32(math.Max(2, float64(gaugeSize)*0.015))
	r.needle.StrokeColor = color.White

	// Position center cap
	capSize := float32(math.Max(10, float64(gaugeSize)*0.1))
	r.centerCap.Resize(fyne.NewSize(capSize, capSize))
	r.centerCap.Move(fyne.NewPos(centerX-capSize/2, centerY-capSize/2))

	// Position value text (below the center for horizontal gauge)
	r.valueText.Text = fmt.Sprintf("%.1f%s", currentValue, config.Unit)
	r.valueText.TextSize = float32(math.Max(12, float64(gaugeSize)*0.12))
	r.valueText.Refresh()
	valueTextSize := fyne.MeasureText(r.valueText.Text, r.valueText.TextSize, r.valueText.TextStyle)
	r.valueText.Move(fyne.NewPos(centerX-valueTextSize.Width/2, centerY+capSize/2+5))

	// Position label text (below value)
	r.labelText.Text = config.Label
	r.labelText.TextSize = float32(math.Max(10, float64(gaugeSize)*0.09))
	r.labelText.Refresh()
	labelTextSize := fyne.MeasureText(r.labelText.Text, r.labelText.TextSize, r.labelText.TextStyle)
	r.labelText.Move(fyne.NewPos(centerX-labelTextSize.Width/2, centerY+capSize/2+25))

	r.lastSize = size
	r.lastValue = currentValue
}

// getColorForValue returns the appropriate color based on thresholds
func (r *gaugeRenderer) getColorForValue(valuePercent float64, config GaugeConfig) color.Color {
	for _, zone := range config.Thresholds {
		if valuePercent >= zone.Start && valuePercent < zone.End {
			return zone.Color
		}
	}
	// Default to last threshold color or arc color
	if len(config.Thresholds) > 0 {
		return config.Thresholds[len(config.Thresholds)-1].Color
	}
	return config.ArcColor
}

// MinSize returns the minimum size for the renderer
func (r *gaugeRenderer) MinSize() fyne.Size {
	return fyne.NewSize(150, 150)
}

// Refresh updates the renderer
func (r *gaugeRenderer) Refresh() {
	r.gauge.mu.RLock()
	config := r.gauge.config
	r.gauge.mu.RUnlock()

	r.labelText.Text = config.Label
	r.Layout(r.lastSize)

	// Refresh all canvas objects
	for _, seg := range r.arcBg {
		canvas.Refresh(seg)
	}
	for _, seg := range r.arcFg {
		canvas.Refresh(seg)
	}
	canvas.Refresh(r.needle)
	canvas.Refresh(r.centerCap)
	canvas.Refresh(r.valueText)
	canvas.Refresh(r.labelText)
}

// Objects returns all canvas objects for rendering
func (r *gaugeRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, arcSegments*2+4)

	// Add background arc segments
	for _, seg := range r.arcBg {
		objects = append(objects, seg)
	}

	// Add foreground arc segments
	for _, seg := range r.arcFg {
		objects = append(objects, seg)
	}

	// Add needle, cap, and text
	objects = append(objects, r.needle, r.centerCap, r.valueText, r.labelText)

	return objects
}

// Destroy cleans up resources
func (r *gaugeRenderer) Destroy() {
	// Stop animation if running
	r.gauge.StopAnimation()
}

// ================== Convenience Functions ==================

// CreateGaugeDisplay creates a gauge wrapped in a container (for compatibility)
func CreateGaugeDisplay(label string, value, maxValue float64, unit string) fyne.CanvasObject {
	config := DefaultGaugeConfig()
	config.Label = label
	config.Max = maxValue
	config.Unit = unit

	gauge := NewGaugeWithConfig(config)
	gauge.SetValue(value)

	return gauge
}

// CreateSpeedometerGauge creates a speedometer-style gauge (alias for compatibility)
func CreateSpeedometerGauge(label string, value float64, unit string) fyne.CanvasObject {
	config := DefaultGaugeConfig()
	config.Label = label
	config.Unit = unit

	gauge := NewGaugeWithConfig(config)
	gauge.SetValue(value)

	return gauge
}

// CreateCompactGauge creates a compact horizontal progress bar gauge
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

// ================== Gauge Panel for Multiple Gauges ==================

// GaugePanel holds multiple gauges for a monitoring panel
type GaugePanel struct {
	widget.BaseWidget

	gauges map[string]*Gauge
	mu     sync.RWMutex
}

// NewGaugePanel creates a new gauge panel
func NewGaugePanel() *GaugePanel {
	p := &GaugePanel{
		gauges: make(map[string]*Gauge),
	}
	p.ExtendBaseWidget(p)
	return p
}

// AddGauge adds a gauge to the panel
func (p *GaugePanel) AddGauge(id, label string) *Gauge {
	p.mu.Lock()
	defer p.mu.Unlock()

	gauge := NewGauge(label)
	p.gauges[id] = gauge
	return gauge
}

// GetGauge returns a gauge by ID
func (p *GaugePanel) GetGauge(id string) *Gauge {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.gauges[id]
}

// SetValue sets the value of a gauge by ID
func (p *GaugePanel) SetValue(id string, value float64) {
	p.mu.RLock()
	gauge, ok := p.gauges[id]
	p.mu.RUnlock()

	if ok {
		gauge.SetValue(value)
	}
}

// CreateRenderer implements fyne.Widget
func (p *GaugePanel) CreateRenderer() fyne.WidgetRenderer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	objects := make([]fyne.CanvasObject, 0, len(p.gauges))
	for _, gauge := range p.gauges {
		objects = append(objects, gauge)
	}

	grid := container.NewGridWithColumns(len(objects))
	for _, obj := range objects {
		grid.Add(obj)
	}

	return widget.NewSimpleRenderer(grid)
}
