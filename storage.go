package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	storageMutex sync.RWMutex
	exams        map[string]Exam
)

func InitStorage() error {
	// Create necessary directories
	dirs := []string{"data", "data/curriculum"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Initialize JSON files if they don't exist
	files := []string{"courses.json", "exams.json", "results.json"}
	for _, file := range files {
		path := filepath.Join("data", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := ioutil.WriteFile(path, []byte("[]"), 0644); err != nil {
				return err
			}
		}
	}

	exams = make(map[string]Exam)

	if data, err := ioutil.ReadFile("data/exams.json"); err == nil {
		var list []Exam
		if err := json.Unmarshal(data, &list); err == nil {
			for _, e := range list {
				exams[e.ID] = e
			}
		}
	}

	loadInitialCourses()
	loadInitialCurriculums()
	return nil
}

func loadInitialCourses() {
	courses, _ := GetAllCourses()
	if len(courses) == 0 {
		initial := []Course{
			{ID: "CS101", Name: "Programming Essentials in C", Description: "Programming basics, algorithms, problem solving"},
			{ID: "CS201", Name: "MS Office", Description: "Word processing, spreadsheet analysis, document automation"},
			{ID: "CS301", Name: "Cybersecurity", Description: "Cybersecurity protection, data privacy, cryptographic encryption"},
			{ID: "CS302", Name: "Technical English", Description: "Technical terminology, professional workplace communication, and IT-related language proficiency."},
		}
		SaveCourses(initial)
	}
}

func loadInitialCurriculums() {
	// Define the map clearly with struct literals
	curriculums := map[string][]CurriculumItem{
		"CS101": {
			{Topic: "Part 1: Problem-Solving Fundamentals", Content: "Problem-solving involves identifying, analyzing, and resolving issues through systematic steps: Identify, Analyze, Generate Solutions, Evaluate/Select, Implement, and Evaluate Outcome. Techniques include Root Cause Analysis (RCA), Brainstorming, 5 Whys, SWOT, and Trial/Error. Logic is visualized via Flowcharts (Oval: Start/End, Parallelogram: I/O, Rectangle: Process, Diamond: Decision). Algorithms are finite, step-by-step instructions characterized by definiteness and effectiveness. Common mistake: Confusing the algorithm (the logic plan) with the program (the code implementation)."},
			{Topic: "Part 2: Introduction to C Programming", Content: "C is a high-level, structured, and modular language. Programming paradigms include Structured Programming (functions/modules) vs. Object-Oriented Programming (objects/encapsulation). The development lifecycle involves writing code, compiling (Syntax Analysis), and executing (Runtime). C programs start at the `main()` function. Key rules: Every statement ends in a semicolon (;); the compiler ignores white space. Common mistake: Forgetting the semicolon or confusing a Compiler (entire code) with an Interpreter (line-by-line). Example: `#include <stdio.h>` is the header for standard I/O."},
			{Topic: "Part 3: Data Types, Operators, and Conversion", Content: "Basic types: `char` (1 byte), `int` (2/4 bytes), `float` (4 bytes), `double` (8 bytes). Modifiers like `unsigned` extend positive ranges. Type conversion includes Implicit (automatic) and Explicit (casting, e.g., `(float)x`). Operators: Arithmetic (+, -, *, /, %), Relational (==, !=, >), Logical (&&, ||, !), and Bitwise (&, |, ^). Prefix/Postfix increment/decrement have different timing impacts. Common mistake: Integer division (5/2 = 2) vs. float division (5.0/2 = 2.5), and using `=` (assignment) instead of `==` (comparison). Example: `const int MIN = 0;` declares a read-only constant"},
			{Topic: "Part 4: Control Statements and User Input", Content: "User input is handled via `scanf(\"%\format\", &variable);`, where the address-of operator (&) is vital. Control logic uses `if`, `else if`, `else`, and the ternary operator `? :` for shortcuts. `switch` statements manage multi-way branching using integer/char constants. Common mistake: Missing the `break` statement in a `switch`, which triggers 'fall-through' (executing all subsequent cases). Example: A Travel Ticket Fare Calculator uses `switch` for transportation types (Bus/Train/Airplane) to apply specific rates per km."},
			{Topic: "Part 5: Iteration and Looping Structures", Content: "Loops repeat code blocks efficiently: `while` (entry-verified), `do-while` (exit-verified, executes at least once), and `for` (pre-defined iterations). `break` exits the loop entirely, while `continue` skips the current iteration. Common mistake: Creating an 'infinite loop' by failing to update the loop counter or setting an incorrect termination condition.  Example: A `for` loop to print a table: `for(int i=1; i<=10; i++) printf(\"%d x %d = %d\\n\", num, i, num * i);`."},
			{Topic: "Part 6: Arrays and Memory Representation", Content: "Arrays store homogeneous data in contiguous memory, indexed 0 to N-1. Variable Length Arrays (VLA) allow size definition at runtime. Multi-dimensional arrays (e.g., 2D matrices) function as 'arrays of arrays.' Syntax errors (compile-time) violate language rules, while Semantic errors (runtime) cause incorrect behavior. Common mistake: 'Buffer Overflow' (accessing an index outside the declared size, such as `arr[5]` for an array of size 5).  Example: `int arr[5] = {0};` initializes all 5 elements to zero."},
			{Topic: "Part 7: Structures and Functions", Content: "Structures group heterogeneous data (structs) using the dot (.) operator. Functions encapsulate reusable logic via a Prototype (declaration) and Definition (body). Recursion is a function calling itself, requiring a 'base case' to prevent stack overflow. Common mistake: Attempting to assign strings directly to struct members instead of using `strcpy()`, or forgetting the base case in recursion. Example (Struct): `struct Product { int id; float price; };`. Example (Function): `float toCelsius(float f) { return (f - 32) * 5/9; }`."},
			{Topic: "Part 8: Assessment and Course Structure", Content: "The course is graded out of 150 marks: Final Exam (45), Assignment 1 (30), Assignment 2 (45), and Quizzes/Tasks (30). Success requires mastering C syntax, understanding the distinction between syntax and semantics, and applying problem-solving skills to real-world scenarios. Learning outcomes focus on writing, debugging, and optimizing C programs with a foundation for advanced memory management and modular design."},
		},
		"CS201": {
			{Topic: "Part 1: Microsoft Word 2016 – Interface & Backstage View", Content: "The Word 2016 interface centers on the Ribbon, a panel of tabs (Home, Insert, Design, etc.) containing functional groups like Font and Paragraph. The Backstage View (accessed via the File tab) enables essential file operations: Info (document details), New (blank or template-based), Open, Save, Save As (different name/type/location), Print (with preview), and Close. The Quick Access Toolbar, located above the Ribbon, holds frequently used commands (Save, Undo, Redo) and is fully customizable via the drop-down arrow or 'More Commands' option."},
			{Topic: "Part 2: Document Setup & Text Formatting Fundamentals", Content: "Before typing, configure margins (Layout tab > Margins > preset or Custom Margins), fonts, and line spacing to save editing time later. Character formatting (font, size, color, bold, italic, underline) is applied by selecting text first, then using the Font group on the Home tab or the Font Dialog Box Launcher for advanced effects. Paragraph formatting (alignment, indents, line spacing) uses the Paragraph group or its dialog box. Use Ctrl+B (bold), Ctrl+I (italic), Ctrl+U (underline) for speed."},
			{Topic: "Part 3: Editing, Navigation & Productivity Shortcuts", Content: "The system clipboard temporarily stores copied or cut content (Ctrl+C to copy, Ctrl+X to cut) for pasting elsewhere (Ctrl+V). Find (Home tab > Find) searches for words via the Navigation Pane, while Go To (Find drop-down > Go To) jumps to specific pages, sections, or comments. Key shortcuts: Ctrl+N (new document), Ctrl+O (open), Ctrl+S (save), F12 (Save As), Ctrl+P (print), Ctrl+W (close), Ctrl+Z (undo), Ctrl+A (select all). Always save before printing or closing."},
			{Topic: "Part 4: Mastering Tables – Creation, Editing & Design", Content: "Tables organize text and numbers into rows and columns intersecting as cells. Create a table via Insert tab > Table menu (grid method up to 10x8) or Insert Table command (custom size). Use Table Tools (appears when clicking a table): Layout tab for splitting cells (select cell > Split Cells > rows/columns), merging cells (select > Merge Cells), inserting rows/columns (Insert Above/Below/Left/Right), and deleting (Delete > Cells/Columns/Rows/Table). Design tab offers Table Styles with live preview."},
			{Topic: "Part 5: Pictures & Graphics – Inserting, Styling & Positioning", Content: "Insert pictures from your computer (Insert > Pictures) or online (Insert > Online Pictures > Bing search). After selection, Picture Tools – Format tab appears. Apply Picture Styles (borders, effects), resize proportionally using corner handles (or exact Height/Width values), rotate with the curved arrow, and reposition using Position (preset layouts) or manual drag – but text wrapping must be set (Wrap Text > Square, Tight, etc.) first. Use Lock aspect ratio to maintain proportions."},
			{Topic: "Part 6: Mail Merge – Form Letters & Bulk Documents", Content: "Mail Merge customizes form letters for multiple recipients using three files: main document (your letter), mailing list (Excel, Access, Outlook, or Office list), and merged document (result). Steps: prepare and save letter > Mailings tab > Start Mail Merge > Letters > Select Recipients > Use an Existing List (choose file/sheet, check header row) > Insert Merge Fields (e.g., <<email>>) > Preview Results > Finish & Merge > Print Documents. Don't forget to save after merging."},
			{Topic: "Part 7: Word Web App & Cloud Collaboration", Content: "The free Microsoft Word Web App (via OneDrive or Office.com) lets you create, edit, and share documents from any browser, though with fewer features than the desktop version. Permission levels: Viewer (see only, no edits/suggestions), Commenter (view + add comments/suggestions, no direct edits), Editor (full edit, accept/reject suggestions, share with others). Google Docs offers similar real-time collaboration with automatic saving, version history, and sharing permissions – all free and cloud-based."},
			{Topic: "Part 8: Excel Spreadsheet Essentials – Cells, Rows & Columns", Content: "Excel manages large data amounts in workbooks (entire file) containing worksheets. Grid layout: rows (numbered, up to 1,048,576) and columns (lettered, up to XFD). The intersection is a cell (e.g., B7), holding text, numbers, or formulas. Hold Shift to select multiple cells. Home ribbon's Number panel formats numeric values. Use Formulas ribbon for formula categories, and Data ribbon for Remove Duplicates (no repeated values) and Text to Columns (split text by delimiter)."},
			{Topic: "Part 9: Excel Formulas, Freeze Panes & Auto-Fill", Content: "Formulas start with = and use cell names (e.g., =B7+B8+B9+B10). Auto-fill predicts series: type two cells (e.g., 1,2 or Monday,Tuesday) then drag the fill handle. Freeze Panes (View ribbon) keeps headers visible while scrolling – select cell below and right of area to freeze; cells above and left (excluding selected) stay fixed. Freeze rows, columns, or both. These features automate calculations and navigation in large spreadsheets."},
			{Topic: "Part 10: Computer Input & Output Devices – Complete List", Content: "Input devices: Keyboard (104/108 keys), Mouse (pointing), Joystick (CAD/games), Track Ball (laptop), Light Pen (select/draw), Digitizer (analog to digital), Scanner (image capture), Microphone (sound), MICR (bank cheques), OCR (read printed text), Bar Code Reader (goods/books). Output devices: CRT monitor (cheap but heavy/flickers), LCD (light, low power), LED (slim, bright, long life), Projector, Printers (Laser: fast/quality; Inkjet: high quality/slow), Plotter (vector graphics), Speakers/Headphones (audio)."},
			{Topic: "Part 11: Detailed Excel Formulas – Complete Reference Guide", Content: "All Excel formulas begin with an equal sign (=), which tells Excel that the subsequent characters constitute a calculation. Formulas follow a specific order of operations known as PEMDAS: Parentheses first, then Exponents, followed by Multiplication and Division (left to right), and finally Addition and Subtraction (left to right). Use parentheses to override this order, e.g., =(5+2)*3 equals 21 versus =5+2*3 which equals 11.Cell References in Formulas – Three reference types control how formulas behave when copied: Relative references (A1) change based on the new location; Absolute references ($A$1) stay fixed regardless of where copied; Mixed references ($A1 or A$1) keep either the column or row fixed.Arithmetic Operators – Perform basic math: + (addition), - (subtraction/negation), * (multiplication), / (division), % (percent), and ^ (exponentiation).Comparison Operators – Return TRUE or FALSE: = (equal to), > (greater than), < (less than), >= (greater than or equal to), <= (less than or equal to), <> (not equal to).Text Operators – The ampersand (&) concatenates text strings into one continuous value; e.g., ='North'&'wind' produces 'Northwind'. Additional text functions include UPPER (converts to uppercase), LOWER (converts to lowercase), PROPER (capitalizes first letter of each word), and TRIM (removes extra spaces).Reference Operators – Colon (:) creates a range reference (A1:A10); comma (,) combines multiple references (SUM(B5:B15,D5:D15)); space creates an intersection reference (B7:D7 C6:C8).Statistical Functions – AVERAGE(range) calculates the mean of values. COUNT(range) counts only cells containing numbers, while COUNTA counts all non-blank cells. MAX(range) finds the highest value; MIN(range) finds the lowest value. STDEV(range) calculates standard deviation, measuring how much values vary from the average.SUM and ROUND Functions – SUM is the most frequently used function, totaling values across continuous or non-continuous ranges (e.g., =SUM(A1:A4)). ROUND(number, num_digits) rounds to a specified number of decimal places – unlike cell formatting, this permanently changes the underlying value.Date and Time Functions – DATE(year, month, day) returns a formatted date (e.g., =DATE(2025,3,31) displays 3/31/2025). TIME(hour, minute, second) in 24-hour format returns time in am/pm format; e.g., =TIME(14,44,24) returns 2:44 PM.Logical Functions – The IF function tests a condition and returns one value if TRUE and another if FALSE: =IF(logical_test, value_if_true, value_if_false). Example: =IF(A5<20, 'Amount is less than twenty', 'Amount is more than twenty'). The AND function returns TRUE only if all arguments are TRUE; OR returns TRUE if any argument is TRUE.Conditional Aggregation Functions – SUMIF(range, criteria, sum_range) adds values only when specified criteria are met; e.g., =SUMIF(C3:C235, '>30', D3:D235) sums purchase orders for more than 30 servers. COUNTIF(range, criteria) counts cells meeting specific criteria. Advanced versions SUMIFS and COUNTIFS handle multiple conditions simultaneously.Lookup and Reference Functions – VLOOKUP searches for a value in the first column of a table and returns a value from a specified column. MATCH locates the position of a value within a range. CHOOSE selects a value from a list based on an index number.Financial Functions – Common financial functions include NPER (number of periods), RATE (interest rate), PV (present value), PMT (periodic payment), and FV (future value) – these correspond to the keys on standard financial calculators. For example, PMT calculates loan installment amounts given interest rate, number of periods, and present value.Array and Dynamic Array Functions – BYCOL applies a LAMBDA function to each column and returns an array of results; BYROW does the same for each row. ARRAYTOTEXT converts an array of values into a single text string.Engineering Functions – Convert between number systems: BIN2DEC (binary to decimal), DEC2HEX (decimal to hexadecimal), and DEC2OCT (decimal to octal). Also includes BITAND, BITOR, and BITXOR for bitwise operations, and CONVERT for unit conversions between measurement systems.Tips for Building Formulas – Instead of typing cell references, click the target cell to insert it automatically – this is especially useful when referencing cells across different worksheets or workbooks. Always verify formula results, and use Paste Special > Values to preserve calculated results when removing formulas."},
		},
		"CS301": {
			{Topic: "Part 1: Cybersecurity Essentials – Definition & Scope", Content: "Cybersecurity is the ongoing effort to protect individuals, organizations, and governments from digital attacks. It focuses on safeguarding organizational data, information systems, and online identities from threats like fraud, hackers, and data breaches. Protection operates at three levels: personal (identity, data, devices), organizational (reputation, data, customers), and governmental (national security, economic stability, citizen safety)."},
			{Topic: "Part 2: The Cybersecurity Cube (McCumber Cube)", Content: "A three-dimensional framework for evaluating information security initiatives. Dimension 1 – Foundational Principles: Confidentiality (prevent unauthorized disclosure via encryption, 2FA), Integrity (protect against modification via hashing/checksums), Availability (ensure access for authorized users via backups, updates). Dimension 2 – States of Data: Data at rest (storage on hard drives, USBs), Data in process (processing/updating), Data in transit (transmission between systems). Dimension 3 – Security Measures: Awareness/training/education for users, Technology solutions (firewalls, software), Policy and procedure (administrative controls like incident response plans)."},
			{Topic: "Part 3: Organizational Data Types", Content: "Organizations maintain three primary data types. Transactional data: details relating to buying/selling, production activities, and employment decisions. Financial data: income statements, balance sheets, cash flow statements providing insight into company health. Intellectual property: patents, trademarks, trade secrets, new product plans that provide economic advantage over competitors. Loss of IP can prove disastrous for a company's future."},
			{Topic: "Part 4: Personal Data, Privacy & Tracking", Content: "Personal data includes any information identifying you: name, SSN, driver's license, date of birth, mother's maiden name, photos, messages. Vulnerable records include medical (EHRs, fitness tracker data like heart rate/blood sugar), educational (attendance, disciplinary reports, IEPs), and employment/financial (paychecks, credit rating, bank details). Online activities are tracked by ISPs (may sell data to advertisers), advertisers (monitor shopping habits), search engines and social media (collect gender, geolocation, religion, politics via search histories), and websites (use cookies leaving data trails). Medical identity theft occurs when criminals steal insurance benefits, causing fraudulent procedures to be saved to your permanent medical records."},
			{Topic: "Part 5: Cryptography – Caesar Cipher", Content: "Encryption secures communications by transforming readable text into ciphertext. The Caesar Cipher is a simple substitution cipher shifting letters by a fixed number. Mathematical form: Encryption E_n(x) = (x + n) mod 26, Decryption D_n(x) = (x - n) mod 26, where A=0, B=1, ..., Z=25. Example: 'HELLO' with shift 3 becomes 'KHOOR'. With shift 13 (ROT13), 'HELLO' becomes 'URYYB'. The cipher wraps around after Z."},
			{Topic: "Part 6: Cryptography – Vigenère Cipher", Content: "A polyalphabetic substitution cipher using a keyword repeated cyclically. Encryption uses the Vigenère square (tabula recta): Ei = (Pi + Ki) mod 26. Decryption: Di = (Ei - Ki) mod 26, handling negative results by adding 26. Example: 'ATTACKATDAWN' with key 'LEMON' (repeated LEMONLEMONLE) produces 'LXFOPVEFRNHR'. When the key has one letter, Vigenère becomes equivalent to Caesar Cipher. Modular arithmetic for negatives: -22 mod 26 = 4 (because -22 = 26 × (-2) + 4)."},
			{Topic: "Part 7: System Bugs & Malware", Content: "Bugs are errors in code causing unexpected results. Syntax Errors: structure/grammar mistakes preventing execution. Logic Errors: flawed reasoning producing incorrect outputs. Runtime Errors: occur during execution (memory leaks, division by zero). Security Bugs: vulnerabilities allowing unauthorized access. Malware types include: Viruses (attach to clean files), Worms (self-replicate across networks), Trojans (disguised as legitimate, create backdoors), Spyware (monitor activity secretly), Adware (flood with unwanted ads), Ransomware (encrypt files, demand payment), Cryptojackers (mine cryptocurrency without consent)."},
			{Topic: "Part 8: E-commerce Cybersecurity Risks", Content: "E-commerce sites are prime targets because they store, process, and transfer large amounts of personal/financial data. Major risks: E-skimming (web skimming) – hackers inject malicious JavaScript into checkout pages to steal credit card data in real time through infiltration, code injection, skimming, while the purchase appears normal. Other risks include malware, phishing/smishing, and DDoS attacks. Consumer protection: use digital wallets (Apple Pay, Google Pay, PayPal) with tokenization, enable transaction alerts, use virtual/one-time cards. Merchant protection: PCI DSS 4.0 compliance, Content Security Policy (CSP), Subresource Integrity (SRI), regular updates for CMS and plugins."},
			{Topic: "Part 9: Identity – Offline vs. Online", Content: "Offline identity is your real-life persona presented daily at home, school, or work, including full name, age, address known to family and friends. Online identity is who you are and how you present yourself online, including usernames/aliases and social identity. Simply using the web establishes an online identity – even without social media accounts. Secure username guidelines: avoid full name, address, phone number, email username, birth year, department names, or password clues; avoid reusing 'super-odd' usernames across platforms as this makes tracking easier."},
		},
		"CS302": {
			{Topic: "Part 1: IT Roles & Job Responsibilities", Content: "Understanding IT roles is essential for workplace communication. Database Administrator (DBA) manages and secures databases: ensures performance/integrity/security, performs backups and recovery, controls user access permissions, optimizes queries for speed. Not responsible for hardware installation, game graphics, or front-end code. Project Manager (IT) oversees IT projects and teams: creates timelines and budgets, assigns tasks to developers/testers/analysts, communicates with stakeholders, manages risks and issues. Not responsible for writing low-level code, fixing hardware, or cleaning staff. Systems Administrator manages computer systems and servers: sets up/deletes user accounts, installs/updates operating systems, monitors system performance, configures network settings, manages system backups. Helpdesk Supervisor manages support team and tickets: assigns tickets to technicians, ensures SLAs (Service Level Agreements) are met, trains support staff, escalates complex issues. Not responsible for programming, motherboard design, or network building (though may oversee). Support Technician assists users with technical issues: answers helpdesk calls/emails, troubleshoots software/hardware problems, resets passwords, guides users through solutions."},
			{Topic: "Part 2: Hardware & Peripherals", Content: "A peripheral is an external device that connects to a computer but is not part of the core (CPU, RAM, motherboard). Examples: monitor (output), mouse (input), keyboard (input), printer (output), scanner (input), speaker (output), external hard drive (storage). NOT peripherals: CPU, RAM, motherboard (internal components). Input devices send data TO the computer: mouse (detects movement/clicks), keyboard (types letters/commands), scanner (digitizes documents), webcam (captures video/images), microphone (captures audio), graphics tablet (uses stylus for drawing/writing). Output devices receive data FROM the computer: monitor (displays visuals), printer (produces physical copies), speaker (produces sound), projector (projects screen onto wall). Touch screen functions as both input AND output. Storage comparison: SSD (Solid State Drive – no moving parts, very fast, moderate capacity), HDD (Hard Disk Drive – spinning platter, slower, very large capacity), USB flash (no moving parts, moderate speed, small-medium capacity), Optical drive (CD/DVD/Blu-ray – spinning disc, slow, small). Key point: SSD is faster and more durable than HDD because it has no moving parts. Connectors: RJ45 (Ethernet), USB-C (modern data/power/display), HDMI (video/audio to monitor/TV), VGA (older analog video). CPU (Central Processing Unit) executes instructions and performs calculations – called the brain of the computer. Other components: RAM (temporary memory, fast but volatile), HDD/SSD (permanent storage, slower), PSU (Power Supply Unit provides power), Motherboard (connects everything). UPS (Uninterruptible Power Supply) provides emergency power during outages, surge protection, time to save work and shut down properly. Not a processing system or speed booster."},
			{Topic: "Part 3: Networking & Internet", Content: "Network: group of linked computers sharing resources. Router: device forwarding data between different networks. Firewall: security system blocking unauthorized access. VPN (Virtual Private Network): secure connection over public networks. Star Topology: all devices connect to central server/switch – if one device fails, others work; if central switch fails, whole network fails. Most common in offices. Other topologies: Bus (single cable, older), Ring (devices in circle), Mesh (every device connects to every other – reliable but expensive). LAN vs WAN: LAN spans one room/building/campus (example: home Wi-Fi, school lab) with very high speed, owned by you/organization. WAN spans cities/countries/continents (example: Internet, branch offices) with lower speed due to distance, owned by multiple organizations/ISPs. Protocols: TCP/IP (Transmission Control Protocol/Internet Protocol) governs all data transmission over internet. HTTPS (HyperText Transfer Protocol Secure) provides secure encrypted web browsing. HTTP provides regular non-secure web browsing. FTP (File Transfer Protocol) transfers files. SMTP (Simple Mail Transfer Protocol) sends emails. Bandwidth: amount of data transferable per month or per second, measured in Mbps (Megabits per second). Higher bandwidth = faster transfer, more visitors. Web concepts: URL (Uniform Resource Locator – web address), Hyperlink (clickable text/image to another page), Cookie (small browser file tracking preferences/login/activity), Cloud computing (internet-based services instead of local hardware). Cloud storage examples: Google Drive, Dropbox, OneDrive, iCloud. NOT cloud storage: External HDD, USB stick, DVD (local/physical)."},
			{Topic: "Part 4: Software & Applications", Content: "Software types: Spreadsheet for numbers/calculations/charts (Excel, Google Sheets). Word processor for writing documents (Word, Google Docs). Database management for organizing structured data (MySQL, Access, Oracle). Web browser for browsing internet (Chrome, Firefox, Edge). Operating System manages hardware and runs software (Windows, macOS, Linux). HTML (HyperText Markup Language) is standard language for web pages using tags like <h1>, <p>, <a>. Not Photoshop, Excel, or Windows. Open Source Software: source code available to view/modify/distribute – not always free but usually is. Examples: Linux, Firefox, VLC, MySQL. Software licensing models: Freemium (basic free, premium paid – Spotify, Zoom), One-time payment (pay once own forever – older Microsoft Office), Subscription (monthly/yearly – Adobe Cloud, Netflix), Ad-supported (free with ads – many mobile games). Patch: small update fixing specific bugs or security vulnerabilities – not a new version, not hardware/cable. Virtual Machine (VM): software emulating a complete computer system – one physical computer runs multiple virtual computers inside it. Uses: testing different OS, server consolidation, running old software. Examples: VMware, VirtualBox. SaaS (Software as a Service): software hosted on cloud accessed via browser – user installs nothing locally. Examples: Google Workspace (Gmail, Docs), Salesforce, Microsoft 365 online."},
			{Topic: "Part 5: Databases", Content: "Database: organized collection of structured data. Table: data organized in rows and columns. Record (row): one complete entry. Field (column): one piece of data (e.g., name, age). Primary Key: unique identifier for each record (Student ID, National ID, email). Properties: every record has different primary key, cannot be NULL (empty), never changes. Query: request to retrieve specific data from database. SQL (Structured Query Language) example: SELECT * FROM Students WHERE Age > 18; retrieves all students older than 18. Common database software: MySQL (open source, very popular), Microsoft Access (small business), Oracle (large enterprise), PostgreSQL."},
			{Topic: "Part 6: Security & Best Practices", Content: "Essential security practices: Use strong unique passwords prevents brute force and credential stuffing. Enable multifactor authentication – even if password stolen, attacker cannot login. Keep software updated patches security holes. Use antivirus detects and removes malware. Do not share login details maintains accountability. Back up data recovers from ransomware or hardware failure. Strong password requirements: at least 12 characters, mix of uppercase/lowercase/numbers/symbols, not dictionary word, not personal info. Example: T7$mKp@9!qLw (not password123). Multifactor Authentication (MFA): using more than one verification method. Factors: something you KNOW (password), something you HAVE (phone/security key), something you ARE (fingerprint/face scan). Example: enter password THEN enter code texted to phone. Malware types: Trojan (disguised as legitimate, does malicious actions), Virus (attaches to files and spreads), Ransomware (encrypts files, demands payment), Spyware (secretly monitors activity). Firewall: Hardware firewall (standalone at network perimeter), Software firewall (program on individual computer). Blocks unauthorized incoming connections and suspicious outgoing traffic. Does not increase CPU speed, edit spreadsheets, or browse web."},
			{Topic: "Part 7: Business & Communication in IT", Content: "Business English phrases: Launch a product (start selling new product to public), Maintain a server (keep server running with updates/security/repairs), Back up data (make copies to prevent loss), Diagnose a problem (find root cause of issue). Email fields: To (main recipients), Cc – Carbon Copy (secondary recipients visible to all), Bcc – Blind Carbon Copy (recipients hidden from others). Use Bcc for mass emails and privacy protection. Video conferencing main benefit: reduces travel costs (no flights/hotels/gas). Other benefits: saves time, allows remote work, records meetings. Company profiles (quiz context): Digital World (hardware manufacturer – fictional), HostElite (hosting company), IBGroup, Futachiba (other fictional names)."},
			{Topic: "Part 8: English Grammar – Adverbs of Frequency & Modal Verbs", Content: "Adverbs of frequency (how often): Always = 100%, Usually = 90%, Normally = 80-90%, Occasionally = 30-40%, Hardly ever = 5-10% (means 'almost never'), Rarely = 10-20%, Never = 0%. Frequency phrases: Weekly = once a week, Daily = every day, Monthly = once a month, Annually = once a year. Modal verbs for rules/suggestions/permission: mustn't = strong rule prohibited ('You mustn't share your password'), don't have to = not required ('You don't have to work on Sunday'), could = suggestion/possibility ('We could buy tablets instead'), should = recommendation ('You should update your software'), must = strong obligation ('You must log out when finished'), can = permission/ability ('You can use the printer'). Polite requests use 'Could you tell me...?' or 'Could you just...?' – not imperatives, shouting, or past tense. Making suggestions: 'Shall we buy new monitors?', 'Could we try a different approach?', 'Why don't we upgrade the server?', 'Let's install the updates tonight.'"},
			{Topic: "Part 9: English Grammar – Verb Tenses & Conditionals", Content: "Verb tenses: Present simple for regular actions/facts/habits ('The server runs 24/7'). Present continuous for actions happening now ('I am installing the software now'). Past simple for completed past actions ('She fixed the bug yesterday'). Future for actions that will happen ('We will deploy the patch tomorrow'). Conditional sentences: Zero Conditional (general truth/scientific fact): If + present simple, present simple – 'If you drop a tablet, it breaks' (always true, no exception). First Conditional (real possibility in future): If + present simple, will + verb – 'If he comes, we will talk.' Second Conditional (unreal/imaginary): If + past simple, would + verb – 'If I had money, I'd buy a server.'"},
			{Topic: "Part 10: Acronyms & Abbreviations – Complete List", Content: "Computer startup: BIOS (Basic Input Output System). Memory: RAM (Random Access Memory). User interface: GUI (Graphical User Interface). Web: URL (Uniform Resource Locator). Document scanning: OCR (Optical Character Recognition). Business/Finance: TCO (Total Cost of Ownership). Hardware/power: UPS (Uninterruptible Power Supply). Device ID: ESN (Electronic Serial Number). Storage interface: SATA (Serial Advanced Technology Attachment). Network speed: Mbps (Megabits per second). Processor: CPU (Central Processing Unit). Storage: HDD (Hard Disk Drive), SSD (Solid State Drive). Hardware: PSU (Power Supply Unit). Web development: HTML (HyperText Markup Language). Web security: HTTPS (HyperText Transfer Protocol Secure). File transfer: FTP (File Transfer Protocol). Email: SMTP (Simple Mail Transfer Protocol). Networking: TCP/IP (Transmission Control Protocol/Internet Protocol). Network security: VPN (Virtual Private Network). Cloud computing: SaaS (Software as a Service). Virtualization: VM (Virtual Machine). Databases: SQL (Structured Query Language). Security: MFA (Multifactor Authentication)."},
			{Topic: "Part 11: Miscellaneous – Cloud, Firewall, Router vs Switch, Task Manager, Storage", Content: "Cloud computing three service models: IaaS (Infrastructure as a Service – rent servers/storage like AWS EC2), PaaS (Platform as a Service – rent development environment like Google App Engine), SaaS (Software as a Service – rent software like Gmail, Office 365). Firewall operation: inspects each data packet, compares against allow/deny rules. Stateful inspection tracks connection state. Next-gen firewall also examines application data. Router vs Switch: Router connects different networks (home network to internet). Switch connects devices within same network. Task Manager (Windows): shortcut Ctrl+Shift+Esc (direct). Alternative: Ctrl+Alt+Del then select Task Manager. Win+R opens Run dialog (not Task Manager). Alt+F4 closes current window. Storage capacity comparison: 4 TB HDD = 4,000 GB (largest), 256 GB SSD = 256 GB, 16 GB RAM = 16 GB temporary memory, DVD = 4.7 GB single layer. Popular operating systems: Linux (open source, free), Windows (proprietary, paid), macOS (proprietary, free with Apple hardware), iOS (mobile Apple, free), Android (open source based, free)."},
			{Topic: "Part 12: Final Exam Tips", Content: "Tip 1: For security questions (75-250), always choose 'Use strong unique passwords' unless another option is clearly correct. Tip 2: 'Almost never' adverb = Hardly ever (not rarely, occasionally, or normally). Tip 3: Database primary key = unique identifier for each record. Tip 4: Zero conditional describes general truths (if + present, present). Tip 5: Polite requests use 'Could you...' not imperatives. Tip 6: Router connects networks; Switch connects devices within network. Tip 7: BIOS starts when computer turns on before operating system. Tip 8: Firewall blocks unauthorized access (like security guard for network). Tip 9: SSD has no moving parts (unlike HDD which spins). Tip 10: Cloud = internet-based (Google Drive, not USB stick)."},
},
}

	for courseID, items := range curriculums {
		path := filepath.Join("data", "curriculum", courseID+".json")

		data, _ := json.MarshalIndent(items, "", "  ")
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			fmt.Printf("Error writing curriculum for %s: %v\n", courseID, err)
		} else {
			fmt.Printf("Successfully wrote curriculum for %s\n", courseID)
		}
	}
} // End of function

func GetAllCourses() ([]Course, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()
	data, err := ioutil.ReadFile("data/courses.json")
	if err != nil {
		return nil, err
	}
	var courses []Course
	if err := json.Unmarshal(data, &courses); err != nil {
		return nil, err
	}
	return courses, nil
}

func SaveCourses(courses []Course) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()
	data, _ := json.MarshalIndent(courses, "", "  ")
	return ioutil.WriteFile("data/courses.json", data, 0644)
}

func GetCurriculum(courseID string) ([]CurriculumItem, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	// Try multiple possible paths
	paths := []string{
		filepath.Join("data", "curriculum", courseID+".json"),
		filepath.Join("data", "curriculum", strings.ToUpper(courseID)+".json"),
		filepath.Join("curriculum", courseID+".json"),
	}

	var data []byte
	var err error
	for _, path := range paths {
		data, err = ioutil.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Return default curriculum if file doesn't exist
		return getDefaultCurriculum(courseID), nil
	}

	var items []CurriculumItem
	if err := json.Unmarshal(data, &items); err != nil {
		return getDefaultCurriculum(courseID), nil
	}

	return items, nil
}

func getDefaultCurriculum(courseID string) []CurriculumItem {
	defaults := map[string][]CurriculumItem{
		"CS101": {
			{Topic: "C Programming Basics", Content: "Variables store data in memory. Types: int, float, char, double. Variables must be declared before use."},
			{Topic: "Control Flow", Content: "If-else statements and loops (for, while) control program flow based on conditions."},
			{Topic: "Functions", Content: "Reusable code blocks that take parameters and return values. Main() is program entry point."},
			{Topic: "Arrays", Content: "Store multiple values of same type in contiguous memory. Indexed from 0."},
		},
		"CS201": {
			{Topic: "Word Processing", Content: "Microsoft Word features: formatting, templates, mail merge, track changes, tables."},
			{Topic: "Spreadsheets", Content: "Excel functions: SUM, VLOOKUP, IF. Pivot tables for data analysis."},
			{Topic: "Automation", Content: "Macros and VBA for automating repetitive tasks in Office."},
		},
		"CS301": {
			{Topic: "Cybersecurity Basics", Content: "CIA triad: Confidentiality, Integrity, Availability. Core security principles."},
			{Topic: "Encryption", Content: "Symmetric (one key) vs Asymmetric (public/private keys). AES, RSA algorithms."},
			{Topic: "Network Security", Content: "Firewalls, IDS/IPS, VPNs, SSL/TLS for secure communication."},
		},
	}

	if curriculum, exists := defaults[courseID]; exists {
		return curriculum
	}

	// Return generic curriculum
	return []CurriculumItem{
		{Topic: "Course Introduction", Content: "Welcome to " + courseID + ". This course covers essential topics in ICT."},
		{Topic: "Key Concepts", Content: "Understanding fundamental principles and applications in this subject area."},
		{Topic: "Practical Applications", Content: "Real-world applications and case studies relevant to this course."},
	}
}

func StoreExam(exam Exam) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()
	exam.CreatedAt = time.Now().Format(time.RFC3339)
	exams[exam.ID] = exam

	var all []Exam
	for _, e := range exams {
		all = append(all, e)
	}
	data, _ := json.MarshalIndent(all, "", "  ")
	return ioutil.WriteFile("data/exams.json", data, 0644)
}

func StoreResult(result ExamResult) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()
	var results []ExamResult
	data, _ := ioutil.ReadFile("data/results.json")
	json.Unmarshal(data, &results)
	results = append(results, result)
	newData, _ := json.MarshalIndent(results, "", "  ")
	return ioutil.WriteFile("data/results.json", newData, 0644)
}
