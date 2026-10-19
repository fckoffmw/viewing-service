import { Link } from 'react-router'

const AuthSwitch = ({to, question, text}) => {
    return (
        <p className="auth-switch">{question} <Link to={to}>{text}</Link></p>
    )
   
}

export default AuthSwitch