export const isDev = () => {
    const development = !process.env.NODE_ENV || process.env.NODE_ENV === 'development'
    return development
}

export const sleep = async (secs) => {
    return new Promise( (resolve, _) => {
        setTimeout(() => {
            resolve(true)
        }, secs * 1000)
    })
}

export const duration = (microsecs) => {
    let sign = ""    
    if (microsecs < 0) {
        microsecs *= -1
        sign = "-"
    }

    let nanosecs = microsecs / 1000
    let msecs = nanosecs / 1000

    let secs = msecs / 1000
    let minutes = Math.floor(secs / 60)
    let hours = Math.floor(minutes / 60)

    if (secs >= 60) {
        secs %= 60
    }
    if (minutes >= 60) {
        minutes %= 60
    }
    let out = sign

    if (hours !== 0) out = `${out}${hours}h`
    if (minutes !== 0) out = `${out}${minutes}m`
    if (secs !== 0) out = `${out}${secs}s`

    return out
}