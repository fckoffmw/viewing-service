import { useEffect } from 'react'

const RedirectToRooms = () => {
    useEffect(() => {
        window.location.replace('/rooms.html')
    }, [])

    return null
}

export default RedirectToRooms