import { useEffect, useState } from "react"
import useRequireAuth from "../hooks/useRequireAuth"
import RoomItem from "../components/RoomItem"
import SakuraFall from "../components/SakuraFall"

const RoomsPage = () => {
    const {id} = useRequireAuth()

    const [rooms, setRooms] = useState([])
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState('')

    useEffect(() => {
        const loadRooms = async () => {
            try {
                const res = await fetch('/api/rooms')
                if (!res.ok) {
                    setError('Ошибка загрузки')
                    return
                }
                setRooms(await res.json())
            } catch {
                setError('Ошибка соединения')
            } finally {
                setIsLoading(false)
            }
        }
        loadRooms()
    }, [])

    if (isLoading) return <p className="empty">Загрузка...</p>
    if (error) return <p className="empty">{error}</p>
    if (rooms.length === 0) return <p className="empty">Нет комнат. Создайте первую</p>


    return (
        <div className="page">
            <SakuraFall />
            {rooms.map((room) => <RoomItem key={room.id} room={room} userID={id} />)}
        </div>
    )
}

export default RoomsPage