const RoomItem = ({ room, userID }) => {
    const source = room.current_source ? room.current_source.name : '\u2014'
    const members = room.members_online || 0
    const isOwner = room.owner_id === userID
    const name = room.name
    const inviteCode = room.invite_code

    return (
        <div className="room-item">
            <div className="room-info">
                <h3>{name}</h3>
                <div className="room-meta"> {source} · {members} онлайн · <span className="room-code" title="Копировать ссылку">{inviteCode}</span></div>
            </div>

            <div className="room-actions">
                <a href={'/room.html?invite=' + inviteCode} className="btn btn-primary btn-sm">Войти</a>
                {isOwner && <button className="btn btn-danger btn-sm btn-delete-room">✕</button>}
            </div>
        </div>
    )
}

export default RoomItem