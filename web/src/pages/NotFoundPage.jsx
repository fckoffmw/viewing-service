import { Link } from 'react-router'
import AuthNav from '../components/AuthNav'
import SakuraFall from '../components/SakuraFall'
import './NotFoundPage.css'

const NotFoundPage = () => {
    return (
        <>
            <SakuraFall />
            <AuthNav />
            <main className="not-found">
                <div className="not-found__card">
                    <p className="not-found__code">404</p>
                    <h1 className="not-found__title">Страница не найдена</h1>
                    <p className="not-found__text">Похоже, её унесло ветром вместе с лепестками сакуры.</p>
                    <div className="not-found__actions">
                        <a href="/" className="btn btn-primary">На главную</a>
                        <Link to="/login" className="btn btn-secondary">Войти</Link>
                    </div>
                </div>
            </main>
        </>
    )
}

export default NotFoundPage
