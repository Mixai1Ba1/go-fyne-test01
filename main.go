package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var content *fyne.Container

var (
	buttons        = make(map[string]*widget.Button)
	highlightedKey string
	startTime      time.Time
	attempts       int
	currentLevel   int = 1
	testRunning    bool
	results        []float64
	graphImage     *canvas.Image
	speedLevels    = map[int]time.Duration{
		1: 3000 * time.Millisecond,
		2: 2500 * time.Millisecond,
		3: 2000 * time.Millisecond,
		4: 1500 * time.Millisecond,
		5: 1000 * time.Millisecond,
	}
	numPadEnabled bool // Флаг включения NumPad
)

// Коды клавиш для цифр (верхний блок и NumPad)
var keyCodes = map[int]string{
	// Верхний блок цифр
	18: "1", 19: "2", 20: "3", 21: "4", 23: "5", 22: "6", 26: "7", 28: "8", 25: "9", 29: "0",
	// NumPad
	83: "Num1", 84: "Num2", 85: "Num3", 86: "Num4", 87: "Num5", 88: "Num6", 89: "Num7", 91: "Num8", 92: "Num9", 82: "Num0",
}

// Функция выбора случайной клавиши в зависимости от уровня
func getNextKey() string {
	switch currentLevel {
	case 1:
		return "1" // Фиксированная клавиша для уровня 1
	case 2:
		// Случайная клавиша на верхней панели
		keys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
		return keys[rand.Intn(len(keys))]
	case 3:
		return "Num1" // Фиксированная клавиша для уровня 3
	case 4:
		// Случайная клавиша на NumPad
		keys := []string{"Num1", "Num2", "Num3", "Num4", "Num5", "Num6", "Num7", "Num8", "Num9", "Num0"}
		return keys[rand.Intn(len(keys))]
	case 5:
		// Случайная клавиша на всех панелях
		keys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "Num1", "Num2", "Num3", "Num4", "Num5", "Num6", "Num7", "Num8", "Num9", "Num0"}
		return keys[rand.Intn(len(keys))]
	default:
		return "1" // По умолчанию
	}
}

// Функция подсветки клавиши
func highlightRandomKey(label *widget.Label, window fyne.Window) {
	if !testRunning {
		return
	}

	// Убираем подсветку с предыдущей клавиши
	if btn, exists := buttons[highlightedKey]; exists {
		btn.Importance = widget.MediumImportance
		btn.Refresh()
	}

	// Выбираем новую клавишу
	highlightedKey = getNextKey()
	if btn, exists := buttons[highlightedKey]; exists {
		btn.Importance = widget.HighImportance
		btn.Refresh()
	}

	// Обновляем текст метки
	label.SetText("Нажмите: " + highlightedKey)

	// Фиксируем время начала реакции
	startTime = time.Now()

	// Обновляем интерфейс
	window.Canvas().Refresh(label)
}

func keyPressed(input string, label *widget.Label, window fyne.Window, content *fyne.Container) {

	if !testRunning {
		return
	}

	if input != highlightedKey {
		label.SetText(fmt.Sprintf("Ошибка: Вы нажали не ту клавишу. Ожидалось: %s", highlightedKey))
		return
	}

	// Фиксируем время реакции
	reactionTime := time.Since(startTime).Seconds()
	results = append(results, reactionTime)
	attempts++

	// Проверяем завершение теста
	if attempts >= 10 {
		testRunning = false
		label.SetText("Тест завершен")
		saveResults()

		return
	}

	highlightRandomKey(label, window)
}

// Функция сохранения результатов
func saveResults() {
	file, err := os.OpenFile("reaction_results.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("\nТест: Уровень %d\n", currentLevel))
	for i, time := range results {
		file.WriteString(fmt.Sprintf("Нажатие %d: %.3f сек\n", i+1, time))
	}
}

// Функция рисования графика
func drawGraph() {
	p := plot.New()
	p.Title.Text = "Время реакции"
	p.X.Label.Text = "Нажатие"
	p.Y.Label.Text = "Время (сек)"

	points := make(plotter.XYs, len(results))
	for i, time := range results {
		points[i].X = float64(i + 1)
		points[i].Y = time
	}

	s, err := plotter.NewScatter(points)
	if err != nil {
		log.Fatal(err)
	}
	p.Add(s)

	err = p.Save(150*vg.Millimeter, 100*vg.Millimeter, "reaction_graph.png")
	if err != nil {
		log.Fatal(err)
	}

	// Обновляем изображение графика в GUI
	graphImage.File = "reaction_graph.png"
	graphImage.Resize(fyne.NewSize(450, 300))
	graphImage.Refresh()
}

// Функция запуска теста
func startTest(label *widget.Label, window fyne.Window) {
	testRunning = true
	attempts = 0
	results = []float64{}
	highlightRandomKey(label, window)
}

// Функция смены уровня
func changeLevel(level int, label *widget.Label, numPadContainer *fyne.Container) {
	currentLevel = level
	testRunning = false
	label.SetText("Выберите уровень и нажмите 'Старт'")

	// Включаем NumPad только для уровней 3, 4 и 5
	numPadEnabled = (level >= 3)
	if numPadEnabled {
		numPadContainer.Show()
	} else {
		numPadContainer.Hide()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	myApp := app.New()
	myWindow := myApp.NewWindow("Тест скорости реакции")
	myWindow.Resize(fyne.NewSize(900, 700))

	// Метка текущей клавиши
	label := widget.NewLabel("Выберите уровень и нажмите 'Старт'")

	// Кнопки 0-9
	buttonGrid := container.NewGridWithColumns(10)
	for _, num := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		btn := widget.NewButton(num, func(n string) func() {
			return func() {
				keyPressed(n, label, myWindow, content)
				drawGraph()
			}
		}(num))
		buttons[num] = btn
		buttonGrid.Add(btn)
	}

	// NumPad (изначально скрыт)
	numPadContainer := container.NewGridWithColumns(3)
	for _, num := range []string{"Num1", "Num2", "Num3", "Num4", "Num5", "Num6", "Num7", "Num8", "Num9", "Num0"} {
		btn := widget.NewButton(num, func(n string) func() {
			return func() {
				keyPressed(n, label, myWindow, content)
				drawGraph()
			}
		}(num))
		buttons[num] = btn
		numPadContainer.Add(btn)
	}
	numPadContainer.Hide()

	// Кнопки уровней
	levelButtons := container.NewHBox()
	for i := 1; i <= 5; i++ {
		lvl := i
		levelButtons.Add(widget.NewButton("Уровень "+strconv.Itoa(lvl), func() {
			changeLevel(lvl, label, numPadContainer)
		}))
	}

	// Кнопка "Старт"
	startButton := widget.NewButton("Старт", func() {
		startTest(label, myWindow)
	})

	// Поле для графика
	graphImage = canvas.NewImageFromFile("reaction_graph.png")
	graphImage.Resize(fyne.NewSize(450, 300))
	graphImage.FillMode = canvas.ImageFillContain

	// Основной контейнер
	content = container.NewVBox(
		label,
		startButton,
		levelButtons,
		buttonGrid,
		numPadContainer,
		graphImage,
	)

	myWindow.SetContent(content)

	// Обработка нажатий клавиш по их кодам
	myWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		keyCode := int(event.Physical.ScanCode)
		fmt.Printf("Нажата клавиша с кодом: %d\n", keyCode) // Отладочное сообщение

		if key, exists := keyCodes[keyCode]; exists {
			keyPressed(key, label, myWindow, content)
		} else {
			label.SetText(fmt.Sprintf("Ошибка: Клавиша с кодом %d не распознана", keyCode))
		}
		drawGraph()
	})

	myWindow.ShowAndRun()
}
