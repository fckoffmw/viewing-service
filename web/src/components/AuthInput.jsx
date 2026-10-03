const AuthInput = (props) => {
    const {
        type,
        id,
        placeholder,
        autoComplete,
        minLength,
        value,
        onChange,
    } = props

    return (
            <input value={value} onChange={onChange} className="auth-input" type={type} id={id} placeholder={placeholder} autoComplete={autoComplete} required minLength={minLength} />
    )
}

export default AuthInput