;(() => {
  
  
  
  let expectingMessage = false
  function dial() {
    const conn = new WebSocket(`ws://${location.host}/subscribe`)

    conn.addEventListener("close", ev => {
      appendLog(`WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`, true)
      if (ev.code !== 1001) {
        appendLog("Reconnecting in 1s", true)
        setTimeout(dial, 1000)
      }
    })
    conn.addEventListener("open", ev => {
      console.info("websocket connected")
    })

    
    conn.addEventListener("message", ev => {
      if (typeof ev.data !== "string") {
        console.error("unexpected message type", typeof ev.data)
        return
      }
      const p = appendLog(ev.data)
      if (expectingMessage) {
        p.scrollIntoView()
        expectingMessage = false
      }
    })
  }
  dial()

  const messageLog = document.getElementById("message-log")
  const publishForm = document.getElementById("publish-form")
  const messageInput = document.getElementById("message-input")

  
  function appendLog(text, error) {
    const p = document.createElement("p")
    
    p.innerText = `${new Date().toLocaleTimeString()}: ${text}`
    if (error) {
      p.style.color = "red"
      p.style.fontStyle = "bold"
    }
    messageLog.append(p)
    return p
  }
  appendLog("Submit a message to get started!")

  
  publishForm.onsubmit = async ev => {
    ev.preventDefault()

    const msg = messageInput.value
    if (msg === "") {
      return
    }
    messageInput.value = ""

    expectingMessage = true
    try {
      const resp = await fetch("/publish", {
        method: "POST",
        body: msg,
      })
      if (resp.status !== 202) {
        throw new Error(`Unexpected HTTP Status ${resp.status} ${resp.statusText}`)
      }
    } catch (err) {
      appendLog(`Publish failed: ${err.message}`, true)
    }
  }
})()
