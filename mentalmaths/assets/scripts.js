
const signin = async (e) => {
    e.preventDefault()
    clearMessages()
    const username = document.getElementById("username").value
    const password = document.getElementById("password").value
    if (!username || !password) return alertMessage(`Invalid fields`, false)

    const r = await axios.post('/signin', {Username: username, Password: password},{withCredentials: true}).then(r => true).catch(e => false)
    if (r === false) return alertMessage(`Username or password are not correct`, false)
    alertMessage(`Sign in success! Redirecting...`, true)
    setTimeout(() => {
        window.location.replace("/");
    },1000)
  
}
const signup = async (e) => {
    e.preventDefault()
    clearMessages()
    const username = document.getElementById("username").value
    const password = document.getElementById("password").value
    if (!username || !password) return alertMessage(`Invalid fields`, false)

    const r = await axios.post('/signup', {Username: username, Password: password},{withCredentials: true}).then(r => true).catch(e => false)
    if (r === false) return alertMessage(`Error creating account`, false)
    alertMessage(`Account created successfully! Redirecting...`, true)
    setTimeout(() => {
        window.location.replace("/signin");
    },1000)
  
}
const beginGame = async (e) => {
    e.preventDefault()
    clearMessages()
    const nquestions = parseInt(document.getElementById("nquestions").value)
    const maxnumber = parseInt(document.getElementById("maxnumber").value)
    if (!nquestions || !maxnumber) return alertMessage(`Invalid fields`, false)
    const r = await axios.post('/api', {QuestionNumber: nquestions, MaxNumber: maxnumber},{withCredentials: true}).then(r => r.data).catch(e => false)
    if (r === false) return alertMessage(`Error creating game`, false)
    startGame(r)
}
const submitGame = async (e) => {
    e.preventDefault()
    clearMessages()
    const id = document.getElementById("id").value
    
    const num = parseInt(document.getElementById("numberq").value)
    const answers = []
    for (let i = 0; i < num; i++) {
        answers.push(parseInt(document.getElementById(`answer${i}`).value))
    }
    const r = await axios.post(`/answers/${id}`, {Answers:answers},{withCredentials: true}).then(r => r.data).catch(e => false)
    if (r === false) return alertMessage(`Error submitting game`, false)
    showResults(r)
}
const loadScores = async () => {
    clearMessages()
    const highscores = document.getElementById("highscores")
    const results = document.getElementById("results")
    const r = await axios.get(`/highscore`,{withCredentials: true}).then(r => r.data).catch(e => false)
    if (r === false) return alertMessage(`Error loading scores`, false)
    if (r === null) return
    const sorted = r.sort().reverse()
    results.innerHTML = ''

    sorted.forEach((h,i) => {
        results.innerHTML += `<b>${i + 1}.</b> ${h} points <br>`
    })
    highscores.style.display = 'block'
}
const logout = async () => {
    clearMessages()
    const r = await axios.get(`/logout`,{withCredentials: true}).then(r => true).catch(e => false)
    if (r === false) return alertMessage(`Error logging out`, false)
    window.location.replace("/signin");
}
const showResults = (score) => {
    const postgame = document.getElementById("postgame")
    const ingame = document.getElementById("ingame")
    ingame.style.display = 'none'
    postgame.style.display = 'block'
    document.getElementById("finalscore").innerText = score
}
const restartGame = async () => {
    const postgame = document.getElementById("postgame")
    const pregame = document.getElementById("pregame")
    await loadScores()
    postgame.style.display = 'none'
    pregame.style.display = 'block'
}
const startGame = (questions) => {
    const pregame = document.getElementById("pregame")
    const ingame = document.getElementById("ingame")
    const questionsdiv = document.getElementById("questions")
    pregame.style.display = 'none'
    ingame.style.display = 'block'
    questionsdiv.innerHTML = ''
    questions.SliceOP.forEach((q,i) => {
        questionsdiv.innerHTML += `<span>${questions.SliceOne[i]}</span> <span>${q}</span> <span>${questions.SliceTwo[i]}</span> = <input style="width:70px" name="answer${i}" id="answer${i}" type="number" required><br>`
    })
    questionsdiv.innerHTML += `<input name="id" id="id" value="${questions.Id}" type="hidden" required>`
    questionsdiv.innerHTML += `<input name="numberq" id="numberq" value="${questions.SliceOne.length}" type="hidden" required>`
}
const alertMessage = (str, positive) => {
    const alertm = document.getElementById(`alert${positive ?'positive':'negative'}`)
    alertm.style.display = 'block'
    alertm.innerText = str
}
const clearMessages = () => {
    const alertp = document.getElementById(`alertpositive`)
    const alertn = document.getElementById(`alertnegative`)
    alertp.style.display = 'none'
    alertn.style.display = 'none'
}
if (window.location.pathname == '/') loadScores()