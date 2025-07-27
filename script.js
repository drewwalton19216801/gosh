// Smooth scrolling for navigation links
document.addEventListener('DOMContentLoaded', function() {
    // Smooth scrolling for anchor links
    const links = document.querySelectorAll('a[href^="#"]');
    
    links.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href').substring(1);
            const targetElement = document.getElementById(targetId);
            
            if (targetElement) {
                const offsetTop = targetElement.offsetTop - 100; // Account for fixed navbar
                
                window.scrollTo({
                    top: offsetTop,
                    behavior: 'smooth'
                });
            }
        });
    });
    
    // Add active class to navigation items based on scroll position
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.nav-menu a[href^="#"]');
    
    function updateActiveNav() {
        let current = '';
        
        sections.forEach(section => {
            const sectionTop = section.offsetTop - 150;
            const sectionHeight = section.offsetHeight;
            
            if (window.scrollY >= sectionTop && window.scrollY < sectionTop + sectionHeight) {
                current = section.getAttribute('id');
            }
        });
        
        navLinks.forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('href') === `#${current}`) {
                link.classList.add('active');
            }
        });
    }
    
    window.addEventListener('scroll', updateActiveNav);
    
    // Terminal typing animation for individual terminals
    const terminals = document.querySelectorAll('.terminal');
    
    terminals.forEach((terminal, terminalIndex) => {
        const terminalLines = terminal.querySelectorAll('.terminal-line');
        let currentLine = 0;
        let animationStarted = false;
        
        function showNextLine() {
            if (currentLine < terminalLines.length) {
                terminalLines[currentLine].style.opacity = '1';
                terminalLines[currentLine].style.transform = 'translateY(0)';
                currentLine++;
                setTimeout(showNextLine, 1200);
            }
        }
        
        // Initialize terminal lines as hidden
        terminalLines.forEach(line => {
            line.style.opacity = '0';
            line.style.transform = 'translateY(10px)';
            line.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
        });
        
        // Start animation when terminal comes into view
        const terminalObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting && !animationStarted) {
                    animationStarted = true;
                    setTimeout(showNextLine, 500 + (terminalIndex * 200));
                    terminalObserver.unobserve(terminal);
                }
            });
        }, {
            threshold: 0.3,
            rootMargin: '0px 0px -100px 0px'
        });
        
        terminalObserver.observe(terminal);
    });
    
    // Legacy support for hero section (for other pages)
    const hero = document.querySelector('.hero');
    if (hero) {
        const heroTerminalLines = hero.querySelectorAll('.terminal-line');
        let currentHeroLine = 0;
        
        function showNextHeroLine() {
            if (currentHeroLine < heroTerminalLines.length) {
                heroTerminalLines[currentHeroLine].style.opacity = '1';
                heroTerminalLines[currentHeroLine].style.transform = 'translateY(0)';
                currentHeroLine++;
                setTimeout(showNextHeroLine, 800);
            }
        }
        
        // Initialize hero terminal lines as hidden
        heroTerminalLines.forEach(line => {
            line.style.opacity = '0';
            line.style.transform = 'translateY(10px)';
            line.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
        });
        
        const heroObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    setTimeout(showNextHeroLine, 1000);
                    heroObserver.unobserve(hero);
                }
            });
        });
        
        heroObserver.observe(hero);
    }
    
    // Add scroll-based animations for feature cards
    const featureCards = document.querySelectorAll('.feature-card, .doc-card');
    
    const cardObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    });
    
    featureCards.forEach(card => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(30px)';
        card.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        cardObserver.observe(card);
    });
    
    // Copy code blocks to clipboard
    const codeBlocks = document.querySelectorAll('.code-block');
    
    codeBlocks.forEach(block => {
        block.style.position = 'relative';
        
        const copyButton = document.createElement('button');
        copyButton.textContent = 'Copy';
        copyButton.className = 'copy-button';
        copyButton.style.cssText = `
            position: absolute;
            top: 0.5rem;
            right: 0.5rem;
            background: rgba(255, 255, 255, 0.1);
            color: white;
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 4px;
            padding: 0.25rem 0.5rem;
            font-size: 0.75rem;
            cursor: pointer;
            opacity: 0;
            transition: opacity 0.2s;
        `;
        
        block.appendChild(copyButton);
        
        block.addEventListener('mouseenter', () => {
            copyButton.style.opacity = '1';
        });
        
        block.addEventListener('mouseleave', () => {
            copyButton.style.opacity = '0';
        });
        
        copyButton.addEventListener('click', async () => {
            const code = block.querySelector('code').textContent;
            
            try {
                await navigator.clipboard.writeText(code);
                copyButton.textContent = 'Copied!';
                setTimeout(() => {
                    copyButton.textContent = 'Copy';
                }, 2000);
            } catch (err) {
                console.error('Failed to copy code:', err);
            }
        });
    });
});

// Add CSS for active navigation state
const style = document.createElement('style');
style.textContent = `
    .nav-menu a.active {
        color: #2563eb;
        position: relative;
    }
    
    .nav-menu a.active::after {
        content: '';
        position: absolute;
        bottom: -0.5rem;
        left: 0;
        right: 0;
        height: 2px;
        background: #2563eb;
        border-radius: 1px;
    }
`;
document.head.appendChild(style);