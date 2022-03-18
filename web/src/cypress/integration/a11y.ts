import { testScreen } from '../support'

function testA11y(): void {
  describe('no detectable a11y violations', () => {
    before(() => {
      // create one of everything, all associated with each other so that every list will be populated
      cy.createService().then((service) => {
        cy.setScheduleTarget().then((scheduleTarget) => {
          cy.createEPStep({
            epID: service.epID,
            targets: [{ id: scheduleTarget.scheduleID, type: 'schedule' }],
          }).then(() => {
            cy.createManyAlerts(3, { serviceID: service.id })
          })
        })
      })
    })

    function testRoute(testName: string, route: string): void {
      it(testName, () => {
        cy.visit(route)
        cy.injectAxe()
        // todo: workaround for https://github.com/component-driven/cypress-axe/issues/64
        // eslint-disable-next-line cypress/no-unnecessary-waiting
        cy.wait(1000)
        cy.checkA11y(undefined, {
          includedImpacts: ['critical'], // only report and assert for critical impact items
        })
      })
    }

    testRoute('alerts list', '/alerts?allServices=1&filter=all')
    testRoute('rotations list', '/rotations')
    testRoute('schedules list', '/schedules')
    testRoute('escalation policies list', '/escalation-policies')
    testRoute('services list', '/services')
    testRoute('users list', '/users')
    testRoute('profile', '/profile')
    // testRoute('wizard', '/wizard') TODO: fix critical failure
    // testRoute('admin config', '/admin/config')
    // testRoute('admin system limits', '/admin/limits')
    testRoute('admin toolbox', '/admin/toolbox')
    // testRoute('admin message logs', '/admin/message-logs')
    testRoute('api docs', '/docs')
  })
}

testScreen('a11y', testA11y, false, true)
