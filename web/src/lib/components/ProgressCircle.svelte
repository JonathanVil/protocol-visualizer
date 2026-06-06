<script>
    export let progress;

    const radius = 12;
    const circumference = 2 * Math.PI * radius
    const color = "#e24b4a"
</script>

<svg width="40" height="40" viewBox="0 0 40 40">
    <circle cx="18" cy="18" r={radius}
            fill="none"
            stroke="rgba(255,255,255,0.08)"
            stroke-width="3"
    />
    <!-- countdown arc -->
    <circle cx="18" cy="18" r={radius}
            fill="none"
            stroke={color}
            stroke-width="3"
            stroke-linecap="round"
            stroke-dasharray={circumference}
            stroke-dashoffset={circumference * (1 - progress)}
            transform="rotate(-90 18 18)"
    />
</svg>