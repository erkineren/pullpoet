package pr

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"pullpoet/internal/ai"
	"pullpoet/internal/git"
	"strings"
)

//go:embed prompt.md
var promptTemplate string

// Generator handles PR description generation
type Generator struct {
	aiClient     ai.Client
	customPrompt string
}

// Result represents the generated PR description
type Result struct {
	Title string
	Body  string
}

// NewGenerator creates a new PR generator
func NewGenerator(aiClient ai.Client, customPrompt string) *Generator {
	return &Generator{
		aiClient:     aiClient,
		customPrompt: customPrompt,
	}
}

// Generate creates a PR description based on the git diff and optional description
func (g *Generator) Generate(gitResult *git.GitResult, issueContext, repoURL, language string, addSignature bool) (*Result, error) {
	fmt.Println("   📝 Building unified AI prompt...")

	prompt, err := g.buildUnifiedPrompt(gitResult, issueContext, repoURL, language)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	fmt.Printf("   ✅ Unified prompt built (%d characters)\n", len(prompt))

	response, err := g.aiClient.GenerateDescription(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	fmt.Println("   🔍 Parsing AI response...")
	result, err := g.parseResponse(response)
	if err != nil {
		return nil, err
	}
	fmt.Println("   ✅ Response parsed successfully")

	// Add pullpoet signature to the end of the PR body only if requested
	if addSignature {
		result.Body = g.addPullpoetSignature(result.Body)
	}

	return result, nil
}

// loadPromptTemplate loads the unified prompt template from the embedded content or custom file
func (g *Generator) loadPromptTemplate() (string, error) {
	// If custom prompt is provided, load it from file
	if g.customPrompt != "" {
		content, err := os.ReadFile(g.customPrompt)
		if err != nil {
			return "", fmt.Errorf("failed to read custom prompt file '%s': %w", g.customPrompt, err)
		}
		return string(content), nil
	}

	// Otherwise use the embedded default prompt
	return promptTemplate, nil
}

// buildUnifiedPrompt constructs the prompt using the unified template
func (g *Generator) buildUnifiedPrompt(gitResult *git.GitResult, issueContext, repoURL, language string) (string, error) {
	// Load the base prompt template
	baseTemplate, err := g.loadPromptTemplate()
	if err != nil {
		return "", err
	}

	var promptBuilder strings.Builder

	// Add language instruction if not English
	if language != "" && language != "en" {
		languageInstruction := g.getLanguageInstruction(language)
		promptBuilder.WriteString(languageInstruction)
		promptBuilder.WriteString("\n\n")
	}

	// Add the base template
	promptBuilder.WriteString(baseTemplate)
	promptBuilder.WriteString("\n\n")

	// Add context section
	contextSection := g.buildContextSection(gitResult, issueContext)
	promptBuilder.WriteString(contextSection)

	// Add git diff section
	diffSection := g.buildDiffSection(gitResult)
	promptBuilder.WriteString(diffSection)

	// Add repository information if available
	if repoURL != "" {
		repoInfo := extractRepoInfo(repoURL)
		if repoInfo != "" {
			promptBuilder.WriteString(fmt.Sprintf("**Repository**: %s\n", repoInfo))
			promptBuilder.WriteString(fmt.Sprintf("**Default Branch**: %s\n\n", gitResult.DefaultBranch))
		}
	}

	// Add final instruction
	promptBuilder.WriteString("**Analyze the above information and create a professional PR description following the JSON format specified above.**")

	return promptBuilder.String(), nil
}

// getLanguageInstruction returns the appropriate language instruction for the prompt
func (g *Generator) getLanguageInstruction(language string) string {
	switch language {
	case "tr":
		return "**DİL TALİMATI**: Lütfen tüm PR başlığını ve açıklamasını Türkçe olarak oluşturun. Teknik terimleri Türkçe karşılıklarıyla kullanın, ancak kod örnekleri ve dosya adları İngilizce kalabilir."
	case "es":
		return "**INSTRUCCIÓN DE IDIOMA**: Por favor, genera todo el título y descripción del PR en español. Usa términos técnicos en español cuando sea posible, pero los ejemplos de código y nombres de archivos pueden permanecer en inglés."
	case "fr":
		return "**INSTRUCTION DE LANGUE**: Veuillez générer tout le titre et la description du PR en français. Utilisez des termes techniques en français quand c'est possible, mais les exemples de code et noms de fichiers peuvent rester en anglais."
	case "de":
		return "**SPRACHANWEISUNG**: Bitte generieren Sie den gesamten PR-Titel und die Beschreibung auf Deutsch. Verwenden Sie technische Begriffe auf Deutsch, wenn möglich, aber Code-Beispiele und Dateinamen können auf Englisch bleiben."
	case "it":
		return "**ISTRUZIONE LINGUISTICA**: Per favore, genera tutto il titolo e la descrizione del PR in italiano. Usa termini tecnici in italiano quando possibile, ma gli esempi di codice e i nomi dei file possono rimanere in inglese."
	case "pt":
		return "**INSTRUÇÃO DE IDIOMA**: Por favor, gere todo o título e descrição do PR em português. Use termos técnicos em português quando possível, mas exemplos de código e nomes de arquivos podem permanecer em inglês."
	case "ru":
		return "**ЯЗЫКОВАЯ ИНСТРУКЦИЯ**: Пожалуйста, создайте весь заголовок и описание PR на русском языке. Используйте технические термины на русском языке, когда это возможно, но примеры кода и имена файлов могут остаться на английском языке."
	case "ja":
		return "**言語指示**: PRのタイトルと説明をすべて日本語で作成してください。可能な限り技術用語は日本語を使用してください。ただし、コード例とファイル名は英語のままにできます。"
	case "ko":
		return "**언어 지시사항**: PR 제목과 설명을 모두 한국어로 작성해 주세요. 가능한 한 기술 용어는 한국어를 사용하되, 코드 예제와 파일명은 영어로 유지할 수 있습니다."
	case "zh":
		return "**语言指示**: 请用中文创建所有PR标题和描述。尽可能使用中文技术术语，但代码示例和文件名可以保持英文。"
	// European languages
	case "nl":
		return "**TAALINSTRUCTIE**: Genereer alstublieft de volledige PR-titel en beschrijving in het Nederlands. Gebruik technische termen in het Nederlands waar mogelijk, maar codevoorbeelden en bestandsnamen kunnen in het Engels blijven."
	case "sv":
		return "**SPRÅKINSTRUKTION**: Vänligen generera hela PR-titeln och beskrivningen på svenska. Använd tekniska termer på svenska när det är möjligt, men kodexempel och filnamn kan förbli på engelska."
	case "no":
		return "**SPRÅKINSTRUKSJON**: Vennligst generer hele PR-tittelen og beskrivelsen på norsk. Bruk tekniske termer på norsk når det er mulig, men kodeeksempler og filnavn kan forbli på engelsk."
	case "da":
		return "**SPROGINSTRUKTION**: Generer venligst hele PR-titlen og beskrivelsen på dansk. Brug tekniske termer på dansk når det er muligt, men kodeeksempler og filnavne kan forblive på engelsk."
	case "fi":
		return "**KIELIOHJE**: Generoi kaikki PR-otsikko ja kuvaus suomeksi. Käytä tekniset termit suomeksi kun mahdollista, mutta koodiesimerkit ja tiedostonimet voivat pysyä englanniksi."
	case "pl":
		return "**INSTRUKCJA JĘZYKOWA**: Proszę wygenerować cały tytuł i opis PR w języku polskim. Używaj terminów technicznych po polsku, gdy to możliwe, ale przykłady kodu i nazwy plików mogą pozostać w języku angielskim."
	case "cs":
		return "**JAZYKOVÁ INSTRUKCE**: Prosím, vygenerujte celý název a popis PR v češtině. Používejte technické termíny v češtině, když je to možné, ale příklady kódu a názvy souborů mohou zůstat v angličtině."
	case "sk":
		return "**JAZYKOVÁ INŠTRUKCIA**: Prosím, vygenerujte celý názov a popis PR v slovenčine. Používajte technické termíny v slovenčine, keď je to možné, ale príklady kódu a názvy súborov môžu zostať v angličtine."
	case "hu":
		return "**NYELVI UTASÍTÁS**: Kérjük, generálja a teljes PR címet és leírást magyar nyelven. Használjon technikai kifejezéseket magyarul, amikor lehetséges, de a kódpéldák és fájlnevek maradhatnak angolul."
	case "ro":
		return "**INSTRUCȚIUNE LINGVISTICĂ**: Vă rugăm să generați întregul titlu și descrierea PR în limba română. Folosiți termeni tehnici în română când este posibil, dar exemplele de cod și numele fișierelor pot rămâne în engleză."
	case "bg":
		return "**ЕЗИКОВА ИНСТРУКЦИЯ**: Моля, генерирайте целия PR заглавие и описание на български език. Използвайте технически термини на български, когато е възможно, но примерите за код и имената на файлове могат да останат на английски."
	case "hr":
		return "**JEZIČNA UPUTA**: Molimo generirajte cijeli PR naslov i opis na hrvatskom jeziku. Koristite tehničke termine na hrvatskom kada je moguće, ali primjeri koda i imena datoteka mogu ostati na engleskom."
	case "sl":
		return "**JEZIKOVNA NAVODILA**: Prosimo, generirajte celoten PR naslov in opis v slovenščini. Uporabite tehnične izraze v slovenščini, ko je mogoče, vendar lahko primeri kode in imena datotek ostanejo v angleščini."
	case "et":
		return "**KEELETUNNISTUS**: Palun genereerige kogu PR pealkiri ja kirjeldus eesti keeles. Kasutage tehnilisi termineid eesti keeles, kui võimalik, kuid koodinäited ja failinimed võivad jääda inglise keelde."
	case "lv":
		return "**VALODAS INSTRUKCIJA**: Lūdzu, ģenerējiet visu PR virsrakstu un aprakstu latviešu valodā. Izmantojiet tehniskos terminus latviešu valodā, kad iespējams, bet koda piemēri un failu nosaukumi var palikt angļu valodā."
	case "lt":
		return "**KALBOS INSTRUKCIJA**: Prašome sugeneruoti visą PR pavadinimą ir aprašymą lietuvių kalba. Naudokite techninius terminus lietuvių kalba, kai įmanoma, bet kodo pavyzdžiai ir failų pavadinimai gali likti anglų kalba."
	case "mt":
		return "**ISTRUZZJONI TAL-LINGWA**: Jekk jogħġbok, ġenera t-titlu kollu tal-PR u d-deskrizzjoni bil-Malti. Uża termini tekniċi bil-Malti meta possibbli, iżda eżempji ta' kodiċi u ismijiet ta' fajls jistgħu jibqgħu bl-Ingliż."
	case "ga":
		return "**TREORACHA TEANGA**: Cuir gach teideal PR agus cur síos ar fáil i nGaeilge, más é do thoil é. Úsáid téarmaí teicniúla i nGaeilge nuair is féidir, ach is féidir samplaí cód agus ainmneacha comhad a fhágáil i mBéarla."
	case "cy":
		return "**CYFARWYDDIAD IAITH**: Os gwelwch yn dda, cynhyrchwch yr holl deitl PR a'r disgrifiad yn y Gymraeg. Defnyddiwch dermau technegol yn y Gymraeg pan fo'n bosibl, ond gall enghreifftiau cod ac enwau ffeiliau aros yn Saesneg."
	case "is":
		return "**TUNGUMÁLA LEIÐBEININGAR**: Vinsamlegast búðu til allan PR titil og lýsingu á íslensku. Notaðu tæknilega hugtök á íslensku þegar mögulegt er, en kóðadæmi og skráarnöfn geta verið á ensku."
	case "mk":
		return "**ЈАЗИЧНА ИНСТРУКЦИЈА**: Ве молиме генерирајте го целиот PR наслов и опис на македонски јазик. Користете технички термини на македонски кога е можно, но примерите за код и имињата на датотеки можат да останат на англиски."
	case "sq":
		return "**UDHËZIM GJUHËSOR**: Ju lutemi gjeneroni të gjithë titullin dhe përshkrimin e PR në shqip. Përdorni terma teknike në shqip kur është e mundur, por shembujt e kodit dhe emrat e skedarëve mund të mbeten në anglisht."
	case "sr":
		return "**ЈЕЗИЧКА ИНСТРУКЦИЈА**: Молимо вас да генеришете цео PR наслов и опис на српском језику. Користите техничке термине на српском када је могуће, али примери кода и имена фајлова могу остати на енглеском."
	case "uk":
		return "**МОВНА ІНСТРУКЦІЯ**: Будь ласка, створіть весь заголовок та опис PR українською мовою. Використовуйте технічні терміни українською мовою, коли це можливо, але приклади коду та назви файлів можу залишитися англійською мовою."
	case "be":
		return "**МОЎНАЯ ІНСТРУКЦЫЯ**: Калі ласка, стварыце ўвесь загаловак і апісанне PR на беларускай мове. Выкарыстоўвайце тэхнічныя тэрміны на беларускай мове, калі гэта магчыма, але прыклады коду і назвы файлаў могуць застацца на англійскай мове."
	case "ka":
		return "**ენობრივი ინსტრუქცია**: გთხოვთ, შექმნათ მთელი PR სათაური და აღწერა ქართულ ენაზე. გამოიყენეთ ტექნიკური ტერმინები ქართულად, როცა შესაძლებელია, მაგრამ კოდის მაგალითები და ფაილების სახელები შეიძლება დარჩეს ინგლისურად."
	case "hy":
		return "**ԼԵԶՎԱԿԱՆ ՀԻՆՍՏՐՈՒԿՑԻԱ**: Խնդրում ենք ստեղծել ամբողջ PR վերնագիրը և նկարագրությունը հայերենով: Օգտագործեք տեխնիկական տերմիններ հայերենով, երբ հնարավոր է, բայց կոդի օրինակները և ֆայլերի անունները կարող են մնալ անգլերենով:"
	case "az":
		return "**DİL TƏLİMATI**: Zəhmət olmasa, bütün PR başlığını və təsvirini Azərbaycan dilində yaradın. Mümkün olduqda texniki terminləri Azərbaycan dilində istifadə edin, amma kod nümunələri və fayl adları İngilis dilində qala bilər."
	case "kk":
		return "**ТІЛ НҰСҚАУЫ**: Өтінемін, PR тақырыбы мен сипаттамасын қазақ тілінде жасаңыз. Мүмкін болған кезде техникалық терминдерді қазақ тілінде қолданыңыз, бірақ код мысалдары мен файл атаулары ағылшын тілінде қала алады."
	case "ky":
		return "**ТИЛ КОЛДОНМОСУ**: Өтүнөмүн, PR баш аталышын жана сүрөттөмөсүн кыргыз тилинде түзүңүз. Мүмкүн болгондо техникалык терминдерди кыргыз тилинде колдонуңуз, бирок код мисалдары жана файл аталыштары англис тилинде кала алат."
	case "uz":
		return "**TIL KO'RSATMALARI**: Iltimos, butun PR sarlavhasi va tavsifini o'zbek tilida yarating. Mumkin bo'lganda texnik terminlarni o'zbek tilida ishlating, lekin kod misollari va fayl nomlari ingliz tilida qolishi mumkin."
	case "tg":
		return "**ДАСТУРИ ЗАБОН**: Лутфан, унвони PR ва тавсифиро ба забони тоҷикӣ эҷод кунед. Вақте ки имконпазир аст, истилоҳоти техникӣро ба забони тоҷикӣ истифода баред, аммо мисолҳои код ва номҳои файлҳо метавонанд ба забони англисӣ бимонанд."
	case "mn":
		return "**ХЭЛНИЙ ЗААВАР**: Та бүх PR гарчиг болон тайлбарыг монгол хэл дээр үүсгэнэ үү. Боломжтой үед техникийн нэр томьёог монгол хэл дээр ашиглана уу, гэхдээ кодын жишээ болон файлын нэрүүд англи хэл дээр үлдэж болно."
	default:
		return fmt.Sprintf("**LANGUAGE INSTRUCTION**: Please generate all PR title and description in %s language. Use technical terms in %s when possible, but code examples and file names can remain in English.", language, language)
	}
}

// buildContextSection creates the context section with issue and commit info
func (g *Generator) buildContextSection(gitResult *git.GitResult, issueContext string) string {
	var contextBuilder strings.Builder

	// Add issue context if provided
	if issueContext != "" {
		contextBuilder.WriteString("## 📋 Issue/Task Context\n\n")
		contextBuilder.WriteString("```\n")
		contextBuilder.WriteString(issueContext)
		contextBuilder.WriteString("\n```\n\n")
	}

	// Add commit information if available
	if len(gitResult.Commits) > 0 {
		contextBuilder.WriteString("## 📝 Commit History\n\n")

		for _, commit := range gitResult.Commits {
			contextBuilder.WriteString(fmt.Sprintf("- **%s**: %s\n", commit.ShortHash, commit.Message))
			contextBuilder.WriteString(fmt.Sprintf("  *By %s on %s*\n", commit.Author, commit.Date.Format("2006-01-02 15:04")))
		}
		contextBuilder.WriteString("\n")
	}

	return contextBuilder.String()
}

// buildDiffSection creates the git diff section
func (g *Generator) buildDiffSection(gitResult *git.GitResult) string {
	var diffBuilder strings.Builder

	diffBuilder.WriteString("## 🔍 Git Diff to Analyze\n\n")
	diffBuilder.WriteString("```diff\n")
	diffBuilder.WriteString(gitResult.Diff)
	diffBuilder.WriteString("\n```\n\n")

	return diffBuilder.String()
}

// extractRepoInfo extracts repository URL information from a git repository URL
// and removes sensitive information like PAT tokens
func extractRepoInfo(repoURL string) string {
	if repoURL == "" {
		return ""
	}

	// Convert SSH to HTTPS format
	if strings.HasPrefix(repoURL, "git@github.com:") {
		repoURL = strings.Replace(repoURL, "git@github.com:", "https://github.com/", 1)
	}

	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Remove PAT (Personal Access Token) credentials from URL
	// Format: https://username:token@github.com/owner/repo
	// or: https://token@github.com/owner/repo
	if strings.Contains(repoURL, "@") {
		// Find the @ symbol and extract everything after it
		atIndex := strings.LastIndex(repoURL, "@")
		if atIndex != -1 {
			// Extract the protocol part (https://)
			protocolEnd := strings.Index(repoURL, "://")
			if protocolEnd != -1 {
				protocol := repoURL[:protocolEnd+3] // includes "://"
				hostAndPath := repoURL[atIndex+1:]  // everything after @
				repoURL = protocol + hostAndPath
			}
		}
	}

	return repoURL
}

// addPullpoetSignature adds a footer indicating the PR was generated by pullpoet
func (g *Generator) addPullpoetSignature(body string) string {
	provider, model := g.aiClient.GetProviderInfo()
	signature := fmt.Sprintf("\n\n---\n\n*🤖 This PR description was generated by [pullpoet](https://github.com/erkineren/pullpoet) using %s (%s) - an AI-powered tool for creating professional pull request descriptions.*", provider, model)
	return body + signature
}

// parseResponse extracts the title and body from the AI response using multiple parsing strategies
func (g *Generator) parseResponse(response string) (*Result, error) {
	fmt.Printf("   📊 AI response length: %d characters\n", len(response))

	response = strings.TrimSpace(response)

	// Clean up markdown headers that might be before JSON blocks
	// Some AI models return: # ```json instead of ```json
	response = strings.ReplaceAll(response, "# ```json", "```json")
	response = strings.ReplaceAll(response, "## ```json", "```json")
	response = strings.ReplaceAll(response, "### ```json", "```json")

	// Try to parse as JSON first (multiple methods)
	var jsonResult struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	// Method 1: Look for JSON block with ```json markers
	if jsonStart := strings.Index(response, "```json"); jsonStart >= 0 {
		jsonStart += len("```json")
		if jsonEnd := strings.Index(response[jsonStart:], "```"); jsonEnd >= 0 {
			jsonStr := strings.TrimSpace(response[jsonStart : jsonStart+jsonEnd])
			err := json.Unmarshal([]byte(jsonStr), &jsonResult)
			if err == nil {
				fmt.Println("   ✅ Successfully parsed JSON from ```json block")
				return &Result{
					Title: cleanTitle(jsonResult.Title),
					Body:  jsonResult.Body,
				}, nil
			}
			fmt.Printf("   ⚠️  Found ```json block but JSON parsing failed: %v\n", err)
		}
	}

	// Method 2: Look for any JSON object
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &jsonResult); err == nil {
			fmt.Println("   ✅ Successfully parsed JSON object")
			return &Result{
				Title: cleanTitle(jsonResult.Title),
				Body:  jsonResult.Body,
			}, nil
		} else {
			fmt.Printf("   ⚠️  Found JSON-like structure but parsing failed: %v\n", err)
			previewLen := 100
			if len(jsonStr) < previewLen {
				previewLen = len(jsonStr)
			}
			fmt.Printf("   📄 Attempted to parse: %s...\n", jsonStr[:previewLen])
		}
	}

	// Method 3: Try to extract from markdown structure
	if strings.Contains(response, "# ") {
		lines := strings.Split(response, "\n")
		var title, body string
		var bodyLines []string
		titleFound := false

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") && !titleFound {
				title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
				titleFound = true
				// Collect everything after the title as body
				if i+1 < len(lines) {
					bodyLines = lines[i+1:]
				}
				break
			}
		}

		if titleFound {
			body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
			fmt.Println("   ✅ Successfully parsed markdown structure")
			return &Result{
				Title: cleanTitle(title),
				Body:  body,
			}, nil
		}
	}

	// Method 4: Legacy TITLE:/BODY: format
	if strings.HasPrefix(response, "TITLE:") {
		lines := strings.Split(response, "\n")
		var title, body string

		bodyStart := false
		var bodyLines []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "TITLE:") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "TITLE:"))
			} else if strings.HasPrefix(line, "BODY:") {
				bodyStart = true
			} else if bodyStart {
				bodyLines = append(bodyLines, line)
			}
		}

		body = strings.TrimSpace(strings.Join(bodyLines, "\n"))

		if title != "" {
			fmt.Println("   ✅ Successfully parsed TITLE:/BODY: format")
			return &Result{
				Title: cleanTitle(title),
				Body:  body,
			}, nil
		}
	}

	// Fallback: extract from unstructured response
	lines := strings.Split(strings.TrimSpace(response), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	// Find the first meaningful line as title
	title := ""
	bodyStartIndex := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && title == "" {
			title = line
			bodyStartIndex = i + 1
			break
		}
	}

	// Clean up title
	title = cleanTitle(title)

	// Get body from remaining lines
	var body string
	if bodyStartIndex < len(lines) {
		bodyLines := lines[bodyStartIndex:]
		body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}

	fmt.Println("   ⚠️  Used fallback parsing method")
	return &Result{
		Title: title,
		Body:  body,
	}, nil
}

// cleanTitle removes common prefixes and limits length
func cleanTitle(title string) string {
	// Remove common prefixes
	prefixes := []string{"Title:", "PR Title:", "Pull Request Title:", "**Title:**", "📋 **Title:**"}
	for _, prefix := range prefixes {
		title = strings.TrimPrefix(title, prefix)
	}

	title = strings.TrimSpace(title)

	// Remove markdown formatting
	title = strings.TrimPrefix(title, "**")
	title = strings.TrimSuffix(title, "**")
	title = strings.TrimSpace(title)

	// Limit title length
	if len(title) > 80 {
		title = title[:77] + "..."
	}

	return title
}
