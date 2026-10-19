import { useState } from 'react'
import './SakuraFall.css'

const PETAL_COUNT = 24

// Случайные параметры для каждого лепестка: где появится, какого размера, как быстро падает
const createPetals = () => {
    const petals = []
    for (let i = 0; i < PETAL_COUNT; i++) {
        petals.push({
            id: i,
            left: Math.random() * 100,
            size: 10 + Math.random() * 15,
            opacity: 0.6 + Math.random() * 0.4,
            duration: 9 + Math.random() * 9,
            // отрицательная задержка: анимация стартует «с середины», лепестки уже в полёте
            delay: -Math.random() * 18,
            sway: 20 + Math.random() * 40,
            spin: (Math.random() < 0.5 ? -1 : 1) * (90 + Math.random() * 270),
        })
    }
    return petals
}

const SakuraFall = () => {
    // createPetals передаётся без скобок: React вызовет её один раз, при первой отрисовке
    const [petals] = useState(createPetals)

    return (
        <div className="sakura-fall" aria-hidden="true">
            {petals.map((petal) => (
                <div
                    key={petal.id}
                    className="sakura-fall__petal"
                    style={{
                        left: petal.left + '%',
                        '--size': petal.size + 'px',
                        '--opacity': petal.opacity,
                        '--duration': petal.duration + 's',
                        '--delay': petal.delay + 's',
                        '--sway': petal.sway + 'px',
                        '--spin': petal.spin + 'deg',
                    }}
                >
                    <div className="sakura-fall__leaf" />
                </div>
            ))}
        </div>
    )
}

export default SakuraFall
