const imagesMain = document.querySelector('#images')
const alertMain = document.querySelector('#pageAlert')
const photoPrismLogin = document.querySelector('#photoprismLogin')
const slideShowSettings = document.querySelector('#slideShowSettings')
let config = {}
let delay = 10 // Seconds to view each slide
let fade = 3 // Seconds the fade between slides lasts
let slideTime = delay * 1000
let slideTimeout = null
let syncInterval = null
let paused = false

function initialize() {
    ajax('GET', '/getconfig').then(result => {
        const data = safeJSON(result)

        console.log(data)

        if (typeof data !== 'object') {
            pageMessage("Error getting config data, using defaults.", 3000, "caution")
        }

        config = data

        if (!data.configured) {
            photoPrismLogin.classList.add('show')
            return
        }

        // Update images every minute
        syncInterval = window.setInterval(() => {
            sync()
        }, 60 * 1000)

        delay = parseInt(config.delay)
        fade = parseInt(config.fade)

        // Sets time it takes to fade from one image to another
        document.documentElement.style.setProperty('--transitionTime', fade + `s`);

        const images = imagesMain.children
        images[0].classList.add('show')

        slideTimeout = setTimeout(() => {
            if (paused) return
            slide()
        }, config.delay * 1000)
    })    
}
initialize()

function slide(rev) {
    clearTimeout(slideTimeout)

    const current = imagesMain.querySelector('.show')
    const index = Array.from(imagesMain.children).indexOf(current)
    const nextI = getNext(index, rev)
    const next = imagesMain.children[nextI]

    current.classList.remove('next')
    current.classList.remove('show')
    current.classList.add('prev')

    next.classList.add('show')
    next.classList.add('next')

    slideTimeout = setTimeout(() => {
        if (paused) return
        slide()
    }, config.delay * 1000)    
}

function getNext(i, rev) {
    const maxI = imagesMain.children.length - 1

    if (rev) {
        if (i <= 0) return maxI
        return i - 1
    }

    if (i >= maxI) return 0
    return i + 1
}

function sync(prompted) {
    clearInterval(syncInterval)
    ajax('GET', '/sync').then(resp => {
        const data = safeJSON(resp)

        if (data.Changed) {
            window.location.reload()
        }
        if (prompted) {
            pageMessage('Images synced successfully!', 3000, 'subtle success')
        }

        syncInterval = setInterval(() => {
            sync()
        }, 60 * 1000)
        
    }).catch(error => {
        console.error(error)
        pageMessage('Error syncing photos', 3000, 'danger')
    })
}

function updateDomImages(images) {
    const currentImages = imagesMain.children
    imagesMain.innerHTML = ''

    for (index in images) {
        const image = images[index]
        const figure = document.createElement('figure')
        const img = document.createElement('img')
        img.src = `/static/images/${image}`
        figure.appendChild(img)
        imagesMain.appendChild(figure)
    }
}

function updateConfig() {
    const data = new FormData()

    data.append('delay', config.delay)
    data.append('fade', config.fade)

    ajax('POST', '/updateconfig', data).then(response => {
        console.log(safeJSON(response))
        pageMessage('Settings saved', 3000, 'success')
        delay = config.delay
        slideTime = delay * 1000
        fade = config.fade

        clearTimeout(slideTimeout)
        slideTimeout = setTimeout(() => {
            if (paused) return
            slide()
        }, slideTime)

    }).catch(error => {
        error && console.error(error)
        pageMessage('Error saving settings.', 3000, 'danger')
    })
}

let alertTimeout = null
let activeMessage = false
let waitLine = []
function pageMessage(msg, duration, type) {
    if (activeMessage) {
        waitLine.push({
            msg,
            duration,
            type
        })
        return
    }
    
    activeMessage = true
    const messageCont = alertMain.querySelector('#alertMessage')

    messageCont.innerHTML = msg
    type ? alertMain.classList = type : null
    alertMain.classList.add('show')
    

    alertTimeout = setTimeout(() => {
        hidePageMessage()
    }, duration)
}
function hidePageMessage() {
    clearTimeout(alertTimeout)
    const messageCont = alertMain.querySelector('#alertMessage')
    alertMain.classList.remove('show')
    // Prevents colors from changing and text from dissapearing before box fades out
    setTimeout(() => {
        messageCont.innerText = ''
        alertMain.classList = ''
        activeMessage = false
        if (waitLine.length > 0) {
            pageMessage(waitLine[0].msg, waitLine[0].duration, waitLine[0].type)
            waitLine.shift()
        }
    }, 250)
    
}
function overrideMessage(msg, duration, type) {
    clearTimeout(alertTimeout)
    activeMessage = true
    waitLine = []

    const messageCont = alertMain.querySelector('#alertMessage')

    messageCont.innerHTML = msg
    type ? alertMain.classList = type : null
    alertMain.classList.add('show')

    alertTimeout = setTimeout(() => {
        hidePageMessage()
    }, duration)

}
function safeJSON(string) {
    try {
        const json = JSON.parse(string)
        return json
    } catch (error) {
        console.error(error)
        return string
    }
}
function ajax(method, url, data) {
    return new Promise(async (resolve, reject) => {
        
        let xhttp = new XMLHttpRequest();

        xhttp.onreadystatechange = function() {
            if (this.readyState == 4 && (this.status == 200 || this.status == 201)) {
                // If response from server is good, return the response
                resolve(this.response)
            } else if(this.readyState == 4) {
                console.error(this)
                reject()
            }
        };

        xhttp.open(method, url);
        
        if(data) {
            xhttp.send(data)
        } else {
            xhttp.send()
        }
    })
}

const detailsForm = document.querySelector('#settingsForm')
detailsForm.addEventListener('submit', event => {
    event.preventDefault()

    const formData = new FormData(detailsForm)

    ajax(detailsForm.method, detailsForm.action, formData).then(result => {
        const data = safeJSON(result)

        config = data

        document.documentElement.style.setProperty('--transitionTime', fade + `s`);

        clearTimeout(slideTimeout)
        slideTimeout = setTimeout(() => {
            if (paused) return
            slide()
        }, config.delay * 1000)

        pageMessage('Settings saved!', 3000, 'success')
    }).catch(() => {
        pageMessage('Error saving details', 10000, 'danger')
    })

})

const loginForm = document.querySelector('#loginForm')
loginForm.addEventListener('submit', event => {
    event.preventDefault()

    loginForm.querySelectorAll('input').forEach(inp => {
        inp.classList.remove('error')
    })

    const formData = new FormData(loginForm)

    ajax('POST', loginForm.action, formData).then(result => {
        const data = safeJSON(result)

        if (!data.configured) {

            for (index in data) {
                const error = data[index]
                console.log(error)
                const field = loginForm.querySelector(`input[name="${error.Field}"]`)
                field.classList.add('error')
            }

            pageMessage('Error logging in', 3000, 'danger')

            return
        }

        window.location.reload()

    }).catch(error => {
        error && console.error(error)
        pageMessage('Error Logging in.', 10000, 'danger')
    })

})

function logout() {
    ajax('GET', '/logout').then(response => {
        pageMessage('You have been logged out', 3000, 'success')

        window.setTimeout(() => {
            window.location.reload()
        }, 3250)
    }).catch(error => {
        error && console.error(error)
        pageMessage('Error logging out', 3000, 'danger')
    })
}

let saveTimeout = null
function adjustDelay() {
    clearTimeout(saveTimeout)
    let del = parseInt(config.delay) + 5

    if (del >= 305) {
        del = 5
    }

    config.delay = del
    overrideMessage(`Slide duration:<br>${config.delay} seconds`, 3000)

    saveTimeout = setTimeout(() => {
        updateConfig()
    }, 3000)
}

// Keys meant to be from ui buttons (temp regular keystrokes)
document.addEventListener('keydown', event => {
    console.log(event)

    // toggle animation settings
    if (event.key == 'd') {
        adjustDelay()
    }
    if (event.key == 'f') {
        
    }
})

// ---VVV--- Admin controls ---VVV---
// document.addEventListener('click', () => {
//     slide()
// })
document.addEventListener('keydown', event => {
    console.log(event)

    // next slide
    if (event.key == 'ArrowRight') {
        slide()
    }
    // prev slide
    if (event.key == 'ArrowLeft') {
        slide(true)
    }

    // toggle animation settings
    if (event.key == 'e') {
        slideShowSettings.classList.toggle('show')
    }
    // Show dummy message
    if (event.key == 'm') {
        pageMessage('This is a message', 3000, 'danger')
    }
    // Refresh images
    if (event.key == 'r') {
        sync(true)
    }
    // logout of photoprism
    if (event.key == 'l') {
        logout()
    }
})