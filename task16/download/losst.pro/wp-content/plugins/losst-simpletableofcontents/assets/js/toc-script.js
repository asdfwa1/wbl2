(function () {
    let overrideActiveItem = null
    let items = [];

    const getOffsetTop = element => {
        let offsetTop = 0;
        while (element) {
            offsetTop += element.offsetTop;
            element = element.offsetParent;
        }
        return offsetTop;
    }
    const getTOCOffsetTop = element => {
        let offsetTop = 0;
        while (element && element.parentNode !== undefined && element.parentNode.tagName !== 'DIV') {
            offsetTop += element.offsetTop;
            element = element.offsetParent;
        }
        return offsetTop;
    }
    const scrollWidgetTOC = function (element) {
        if (!element) {
            return
        }
        const container = document.querySelector("#simple-toc-widget ul")
        const offset = getTOCOffsetTop(element) - (container.offsetHeight / 2)
        container.scrollTo({top: offset, behavior: 'smooth'})
    }

    const updateTOCWidget = function () {
        document.querySelectorAll("#simple-toc-widget li > a").forEach(function (element, index) {
            element.classList.remove('toc-item-active')
        })

        const offset = window.scrollY + (window.innerHeight / 2);
        for (let i = 1; i < items.length; i++) {
            if (items[i - 1].position < offset && items[i].position > offset) {
                document.querySelector('#simple-toc-widget li[data-toc-section="' + items[i - 1].id + '"] > a').classList.add('toc-item-active')
                scrollWidgetTOC(document.querySelector('#simple-toc-widget li[data-toc-section="' + items[i - 1].id + '"]'))

                break;
            } else if (i === items.length - 1 && items[i].position < offset) {
                document.querySelector('#simple-toc-widget li[data-toc-section="' + items[i].id + '"] > a').classList.add('toc-item-active')

                break;
            }
        }
    }

    document.addEventListener("DOMContentLoaded", function () {
        document.querySelectorAll("#simple-toc-widget li").forEach(function (element, index) {
            const id = element.getAttribute("data-toc-section")
            const heading = document.getElementById(id)

            items.push({
                id: id,
                position: getOffsetTop(heading),
            })

            const link = element.querySelector("a");
            link.addEventListener("click", function () {
                overrideActiveItem = true
                document.querySelectorAll("#simple-toc-widget li > a").forEach(function (element, index) {
                    element.classList.remove('toc-item-active')
                })
                document.querySelector('#simple-toc-widget li[data-toc-section="' + id + '"] > a').classList.add('toc-item-active')
            })
        });

        updateTOCWidget()

    }, false);


    function observerCallback(entries, observer) {
        if (overrideActiveItem !== null) {
            overrideActiveItem = null

            return
        }

        entries.forEach(entry => {
            updateTOCWidget()
        });
    }

    document.addEventListener('DOMContentLoaded', function () {
        let marginObserver = new IntersectionObserver(observerCallback, {
            rootMargin: '-50% 0% -50% 0%',
            threshold: 0
        });
        let centerObserver = new IntersectionObserver(observerCallback, {
            rootMargin: '0px',
            threshold: 1
        });

        let target = '[data-toc-heading="true"]';
        document.querySelectorAll(target).forEach((i) => {
            if (i) {
                marginObserver.observe(i);
                centerObserver.observe(i);
            }
        });
    })
})();
