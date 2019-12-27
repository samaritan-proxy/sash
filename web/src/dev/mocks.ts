import { Dependency } from '../models/models'

export const mockDepedencies: Dependency[] = [
  {
    service_name: "service1",
    create_time: "2019-12-3 14:00:00",
    update_time: "2019-12-3 14:00:00",
    dependencies: ["service3", "service4", "service5", "service6"],
  },
  {
    service_name: "service2",
    create_time: "2019-12-3 14:00:00",
    update_time: "2019-12-3 14:00:00",
    dependencies: ["service3", "service4"],
  }
]