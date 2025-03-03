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

var (
	buttons        = make(map[string]*widget.Button)     // словарь кнопок (цифры 0-9)
	highlightedKey string                                // клавиша, которую надо нажать
	startTime      time.Time                             // время начала реакции
	attempts       int                                   // количество попыток (максимум 10)
	currentLevel   int                               = 1 // уровень сложности
	testRunning    bool                                  // флаг "идёт тест"
	results        []float64                             // список времен реакции
	graphImage     *canvas.Image                         // график реакции
	speedLevels    = map[int]time.Duration{
		1: 3000 * time.Millisecond,
		2: 2500 * time.Millisecond,
		3: 2000 * time.Millisecond,
		4: 1500 * time.Millisecond,
		5: 1000 * time.Millisecond,
	}
	numPadEnabled bool // использовать NumPad в сложных уровнях
)

// выбор случайной клавиши
func getNextKey() string {
	keys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

	// добавил NumPad для уровней 4 и 5
	if numPadEnabled {
		numPadKeys := []string{"Num1", "Num2", "Num3", "Num4", "Num5", "Num6", "Num7", "Num8", "Num9", "Num0"}
		keys = append(keys, numPadKeys...)
	}
	return keys[rand.Intn(len(keys))]
}

// для подсветки клавиш
func highlightRandomKey(label *widget.Label, window fyne.Window) {
	if !testRunning {
		return
	}
	//убираем подсветку с предыдущей клавиши
	if btn, exists := buttons[highlightedKey]; exists {
		btn.Importance = widget.MediumImportance
		btn.Refresh()
	}
	//выбираем новую клавишу
	highlightedKey = getNextKey()
	if btn, exists := buttons[highlightedKey]; exists {
		btn.Importance = widget.HighImportance
		btn.Refresh()
	}
	//обновляем текст метки
	label.SetText("Нажмите: " + highlightedKey)
	//фиксируем время начала реакции
	startTime = time.Now()
	//обновляем интерфейс
	window.Canvas().Refresh(label)
}

// функция обработки нажатия клавиши (клавиатура + кнопки)
func keyPressed(input string, label *widget.Label, window fyne.Window) {
	if !testRunning || input != highlightedKey {
		return
	}
	// фиксирую время реакции
	reactionTime := time.Since(startTime).Seconds()
	results = append(results, reactionTime)
	attempts++
	// Проверяю завершение теста 10 нажатий
	if attempts >= 10 {
		testRunning = false
		label.SetText("Тест завершен")
		saveResults()
		drawGraph()
		window.Canvas().Refresh(label)
		return
	}
	highlightRandomKey(label, window)
}

// функция для сейва результатов
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

// график
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

	//обновляем изображение графика в GUI
	graphImage.File = "reaction_graph.png"
	graphImage.Resize(fyne.NewSize(400, 300))
	graphImage.Refresh()
}

// для запуска теста
func startTest(label *widget.Label, window fyne.Window) {
	testRunning = true
	attempts = 0
	results = []float64{}
	highlightRandomKey(label, window)
}

// функция смены уровня
func changeLevel(level int, label *widget.Label, numPadContainer *fyne.Container) {
	currentLevel = level
	testRunning = false
	label.SetText("Выберите уровень и нажмите 'Старт'")

	// Включаем NumPad только для уровней 4 и 5
	numPadEnabled = (level == 4 || level == 5)
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
	myWindow.Resize(fyne.NewSize(800, 700))

	//метка текущей клавиши
	label := widget.NewLabel("Выберите уровень и нажмите 'Старт'")

	//кнопки 0-9
	buttonGrid := container.NewGridWithColumns(10)
	for _, num := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		btn := widget.NewButton(num, func(n string) func() {
			return func() {
				keyPressed(n, label, myWindow)
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
				keyPressed(n, label, myWindow)
			}
		}(num))
		buttons[num] = btn
		numPadContainer.Add(btn)
	}
	numPadContainer.Hide()

	//кнопки уровней
	levelButtons := container.NewHBox()
	for i := 1; i <= 5; i++ {
		lvl := i
		levelButtons.Add(widget.NewButton("Уровень "+strconv.Itoa(lvl), func() {
			changeLevel(lvl, label, numPadContainer)
		}))
	}

	//кнопка "Старт"
	startButton := widget.NewButton("Старт", func() {
		startTest(label, myWindow)
	})

	//поле для графика
	graphImage = canvas.NewImageFromFile("reaction_graph.png")
	graphImage.Resize(fyne.NewSize(400, 300))

	//основной контейнер
	content := container.NewVBox(
		label,
		startButton,
		levelButtons,
		buttonGrid,
		numPadContainer,
		graphImage,
	)

	myWindow.SetContent(content)

	myWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		key := string(event.Name)

		// Возможные названия клавиш на разных системах
		numPadMapping := map[string]string{
			"KP_1": "Num1", "KP_2": "Num2", "KP_3": "Num3",
			"KP_4": "Num4", "KP_5": "Num5", "KP_6": "Num6",
			"KP_7": "Num7", "KP_8": "Num8", "KP_9": "Num9",
			"KP_0": "Num0",

			// Возможные альтернативные названия (зависит от ОС)
			"KP_End": "Num1", "KP_Down": "Num2", "KP_Next": "Num3",
			"KP_Left": "Num4", "KP_Begin": "Num5", "KP_Right": "Num6",
			"KP_Home": "Num7", "KP_Up": "Num8", "KP_Prior": "Num9",
			"KP_Insert": "Num0",
		}

		if mappedKey, exists := numPadMapping[key]; exists {
			key = mappedKey
		}

		fmt.Println("Key pressed:", key) // Логирование для проверки
		keyPressed(key, label, myWindow)
	})

	myWindow.ShowAndRun()
}
